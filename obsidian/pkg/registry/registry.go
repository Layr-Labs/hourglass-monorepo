package registry

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	pb "github.com/hourglass/obsidian/api/proto/registry"
	"github.com/hourglass/obsidian/internal/docker"
	"github.com/hourglass/obsidian/pkg/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedRegistryServiceServer
	
	config       *config.RegistryConfig
	dockerClient *docker.Client
	images       map[string]*Image
	registries   map[string]*Registry
	cache        *ImageCache
	mu           sync.RWMutex
}

type Image struct {
	ID        string
	Digest    string
	Reference string
	Size      int64
	CreatedAt time.Time
	PulledAt  time.Time
	Tags      []string
	Labels    map[string]string
}

type Registry struct {
	Name      string
	Type      pb.RegistryType
	URL       string
	Enabled   bool
	CreatedAt time.Time
	auth      *RegistryAuth
}

type RegistryAuth struct {
	Username string
	Password string
	Token    string
}

type ImageCache struct {
	mu       sync.RWMutex
	images   map[string]*CachedImage
	maxSize  int64
	currSize int64
}

type CachedImage struct {
	Image     *Image
	LastUsed  time.Time
	CacheTime time.Time
}

func NewServer(config *config.RegistryConfig) (*Server, error) {
	dockerClient, err := docker.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	s := &Server{
		config:       config,
		dockerClient: dockerClient,
		images:       make(map[string]*Image),
		registries:   make(map[string]*Registry),
		cache:        newImageCache(parseSize(config.Cache.MaxSize)),
	}

	if err := s.initializeRegistries(); err != nil {
		return nil, fmt.Errorf("failed to initialize registries: %w", err)
	}

	go s.cacheCleanupLoop()

	return s, nil
}

func (s *Server) PullImage(ctx context.Context, req *pb.ImageReference) (*pb.Image, error) {
	registry, err := s.getRegistry(req.RegistryName)
	if err != nil {
		return nil, err
	}

	if !registry.Enabled {
		return nil, status.Errorf(codes.FailedPrecondition, "registry %s is disabled", req.RegistryName)
	}

	cachedImage := s.cache.Get(req.Reference)
	if cachedImage != nil {
		return s.imageToProto(cachedImage.Image), nil
	}

	fullRef := s.normalizeReference(req.Reference, registry)

	if err := s.dockerClient.PullImage(ctx, fullRef); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to pull image: %v", err)
	}

	imageID := s.generateImageID(fullRef)
	digest := s.calculateDigest(fullRef)

	image := &Image{
		ID:        imageID,
		Digest:    digest,
		Reference: fullRef,
		Size:      0,
		CreatedAt: time.Now(),
		PulledAt:  time.Now(),
		Tags:      s.extractTags(fullRef),
		Labels:    make(map[string]string),
	}

	s.mu.Lock()
	s.images[imageID] = image
	s.mu.Unlock()

	s.cache.Put(req.Reference, image)

	if s.config.Security.VulnerabilityScanEnabled {
		go s.scanImageAsync(imageID)
	}

	return s.imageToProto(image), nil
}

func (s *Server) GetImage(ctx context.Context, req *pb.ImageID) (*pb.Image, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	image, ok := s.images[req.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "image not found: %s", req.Id)
	}

	return s.imageToProto(image), nil
}

func (s *Server) ListImages(ctx context.Context, req *pb.ListImagesRequest) (*pb.ImageList, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var images []*pb.Image
	for _, img := range s.images {
		if req.Filter == "" || strings.Contains(img.Reference, req.Filter) {
			images = append(images, s.imageToProto(img))
		}
	}

	pageSize := 50
	if req.PageSize > 0 {
		pageSize = int(req.PageSize)
	}

	start := 0
	if req.PageToken != "" {
		start = s.parsePageToken(req.PageToken)
	}

	end := start + pageSize
	if end > len(images) {
		end = len(images)
	}

	var nextPageToken string
	if end < len(images) {
		nextPageToken = s.generatePageToken(end)
	}

	return &pb.ImageList{
		Images:        images[start:end],
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Server) DeleteImage(ctx context.Context, req *pb.ImageID) (*emptypb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	image, ok := s.images[req.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "image not found: %s", req.Id)
	}

	s.cache.Remove(image.Reference)
	delete(s.images, req.Id)

	return &emptypb.Empty{}, nil
}

func (s *Server) AddRegistry(ctx context.Context, req *pb.RegistryConfig) (*pb.Registry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.registries[req.Name]; exists {
		return nil, status.Errorf(codes.AlreadyExists, "registry %s already exists", req.Name)
	}

	registry := &Registry{
		Name:      req.Name,
		Type:      req.Type,
		URL:       req.Url,
		Enabled:   true,
		CreatedAt: time.Now(),
		auth:      s.createAuth(req),
	}

	s.registries[req.Name] = registry

	return s.registryToProto(registry), nil
}

func (s *Server) UpdateRegistryCredentials(ctx context.Context, req *pb.RegistryCredentials) (*emptypb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	registry, ok := s.registries[req.RegistryName]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "registry not found: %s", req.RegistryName)
	}

	registry.auth = &RegistryAuth{
		Username: req.Username,
		Password: req.Password,
		Token:    req.Token,
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) ScanImage(ctx context.Context, req *pb.ImageID) (*pb.ScanResult, error) {
	s.mu.RLock()
	image, ok := s.images[req.Id]
	s.mu.RUnlock()

	if !ok {
		return nil, status.Errorf(codes.NotFound, "image not found: %s", req.Id)
	}

	vulnerabilities := s.performScan(image)

	result := &pb.ScanResult{
		ImageId:                  req.Id,
		ScannedAt:                timestamppb.Now(),
		CriticalVulnerabilities: 0,
		HighVulnerabilities:     0,
		MediumVulnerabilities:   0,
		LowVulnerabilities:      0,
		Vulnerabilities:         vulnerabilities,
	}

	for _, v := range vulnerabilities {
		switch v.Severity {
		case "CRITICAL":
			result.CriticalVulnerabilities++
		case "HIGH":
			result.HighVulnerabilities++
		case "MEDIUM":
			result.MediumVulnerabilities++
		case "LOW":
			result.LowVulnerabilities++
		}
	}

	return result, nil
}

func (s *Server) GetImagePolicy(ctx context.Context, req *pb.ImageID) (*pb.Policy, error) {
	scanResult, err := s.ScanImage(ctx, req)
	if err != nil {
		return nil, err
	}

	allowed := true
	var violations []string

	if scanResult.CriticalVulnerabilities > int32(s.config.Security.MaxCriticalVulnerabilities) {
		allowed = false
		violations = append(violations, fmt.Sprintf("Too many critical vulnerabilities: %d > %d",
			scanResult.CriticalVulnerabilities, s.config.Security.MaxCriticalVulnerabilities))
	}

	if scanResult.HighVulnerabilities > int32(s.config.Security.MaxHighVulnerabilities) {
		allowed = false
		violations = append(violations, fmt.Sprintf("Too many high vulnerabilities: %d > %d",
			scanResult.HighVulnerabilities, s.config.Security.MaxHighVulnerabilities))
	}

	return &pb.Policy{
		ImageId:     req.Id,
		Allowed:     allowed,
		Violations:  violations,
		EvaluatedAt: timestamppb.Now(),
	}, nil
}

func (s *Server) initializeRegistries() error {
	for _, reg := range s.config.Registries {
		registryType := s.parseRegistryType(reg.Type)
		credSource := s.parseCredentialSource(reg.CredentialSource)

		registry := &Registry{
			Name:      reg.Name,
			Type:      registryType,
			URL:       reg.URL,
			Enabled:   true,
			CreatedAt: time.Now(),
		}

		if credSource == pb.CredentialSource_CREDENTIAL_SOURCE_SECRET {
			registry.auth = &RegistryAuth{}
		}

		s.registries[reg.Name] = registry
	}

	return nil
}

func (s *Server) getRegistry(name string) (*Registry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if name == "" {
		for _, reg := range s.registries {
			return reg, nil
		}
	}

	registry, ok := s.registries[name]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "registry not found: %s", name)
	}

	return registry, nil
}

func (s *Server) normalizeReference(ref string, registry *Registry) string {
	if strings.Contains(ref, "/") {
		return ref
	}

	switch registry.Type {
	case pb.RegistryType_REGISTRY_TYPE_ECR:
		return fmt.Sprintf("%s/%s", registry.URL, ref)
	case pb.RegistryType_REGISTRY_TYPE_GHCR:
		return fmt.Sprintf("ghcr.io/%s", ref)
	case pb.RegistryType_REGISTRY_TYPE_DOCKER_HUB:
		return fmt.Sprintf("docker.io/%s", ref)
	default:
		return ref
	}
}

func (s *Server) generateImageID(reference string) string {
	h := sha256.Sum256([]byte(reference + time.Now().String()))
	return hex.EncodeToString(h[:])[:12]
}

func (s *Server) calculateDigest(reference string) string {
	h := sha256.Sum256([]byte(reference))
	return "sha256:" + hex.EncodeToString(h[:])
}

func (s *Server) extractTags(reference string) []string {
	parts := strings.Split(reference, ":")
	if len(parts) > 1 {
		return []string{parts[len(parts)-1]}
	}
	return []string{"latest"}
}

func (s *Server) scanImageAsync(imageID string) {
	time.Sleep(2 * time.Second)
}

func (s *Server) performScan(image *Image) []*pb.Vulnerability {
	return []*pb.Vulnerability{}
}

func (s *Server) cacheCleanupLoop() {
	ticker := time.NewTicker(s.config.Cache.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.cache.Cleanup(s.config.Cache.TTL)
	}
}

func (s *Server) imageToProto(img *Image) *pb.Image {
	return &pb.Image{
		Id:        img.ID,
		Digest:    img.Digest,
		Reference: img.Reference,
		SizeBytes: img.Size,
		CreatedAt: timestamppb.New(img.CreatedAt),
		PulledAt:  timestamppb.New(img.PulledAt),
		Tags:      img.Tags,
		Labels:    img.Labels,
	}
}

func (s *Server) registryToProto(reg *Registry) *pb.Registry {
	return &pb.Registry{
		Name:      reg.Name,
		Type:      reg.Type,
		Url:       reg.URL,
		Enabled:   reg.Enabled,
		CreatedAt: timestamppb.New(reg.CreatedAt),
	}
}

func (s *Server) createAuth(config *pb.RegistryConfig) *RegistryAuth {
	if config.CredentialSource == pb.CredentialSource_CREDENTIAL_SOURCE_IAM_ROLE {
		return nil
	}
	return &RegistryAuth{}
}

func (s *Server) parseRegistryType(t string) pb.RegistryType {
	switch strings.ToLower(t) {
	case "ecr", "aws-ecr":
		return pb.RegistryType_REGISTRY_TYPE_ECR
	case "ghcr", "github":
		return pb.RegistryType_REGISTRY_TYPE_GHCR
	case "dockerhub", "docker-hub":
		return pb.RegistryType_REGISTRY_TYPE_DOCKER_HUB
	default:
		return pb.RegistryType_REGISTRY_TYPE_UNSPECIFIED
	}
}

func (s *Server) parseCredentialSource(source string) pb.CredentialSource {
	switch strings.ToLower(source) {
	case "iam-role":
		return pb.CredentialSource_CREDENTIAL_SOURCE_IAM_ROLE
	case "secret":
		return pb.CredentialSource_CREDENTIAL_SOURCE_SECRET
	case "env":
		return pb.CredentialSource_CREDENTIAL_SOURCE_ENV
	default:
		return pb.CredentialSource_CREDENTIAL_SOURCE_UNSPECIFIED
	}
}

func (s *Server) parsePageToken(token string) int {
	return 0
}

func (s *Server) generatePageToken(offset int) string {
	return fmt.Sprintf("%d", offset)
}

func parseSize(size string) int64 {
	return 107374182400
}

func newImageCache(maxSize int64) *ImageCache {
	return &ImageCache{
		images:  make(map[string]*CachedImage),
		maxSize: maxSize,
	}
}

func (c *ImageCache) Get(ref string) *CachedImage {
	c.mu.RLock()
	defer c.mu.RUnlock()

	img, ok := c.images[ref]
	if !ok {
		return nil
	}

	img.LastUsed = time.Now()
	return img
}

func (c *ImageCache) Put(ref string, image *Image) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.images[ref] = &CachedImage{
		Image:     image,
		LastUsed:  time.Now(),
		CacheTime: time.Now(),
	}
}

func (c *ImageCache) Remove(ref string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.images, ref)
}

func (c *ImageCache) Cleanup(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for ref, img := range c.images {
		if now.Sub(img.CacheTime) > ttl {
			delete(c.images, ref)
		}
	}
}