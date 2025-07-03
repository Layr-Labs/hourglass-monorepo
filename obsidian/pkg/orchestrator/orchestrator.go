package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	pb "github.com/hourglass/obsidian/api/proto/orchestrator"
	"github.com/hourglass/obsidian/internal/docker"
	"github.com/hourglass/obsidian/pkg/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedOrchestratorServiceServer
	
	config         *config.OrchestratorConfig
	dockerClient   *docker.Client
	containers     map[string]*Container
	tasks          map[string]*Task
	taskQueue      chan *Task
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

type Container struct {
	ID        string
	ImageID   string
	State     pb.ContainerState
	Resources *pb.Resources
	CreatedAt time.Time
	StartedAt time.Time
	Labels    map[string]string
}

type Task struct {
	ID          string
	ContainerID string
	Payload     []byte
	State       pb.TaskState
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
	Result      []byte
	Error       string
	Metrics     *pb.TaskMetrics
	Timeout     time.Duration
	Cancel      context.CancelFunc
}

func NewServer(config *config.OrchestratorConfig) (*Server, error) {
	dockerClient, err := docker.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	s := &Server{
		config:       config,
		dockerClient: dockerClient,
		containers:   make(map[string]*Container),
		tasks:        make(map[string]*Task),
		taskQueue:    make(chan *Task, config.Queue.MaxQueueSize),
		ctx:          ctx,
		cancel:       cancel,
	}

	s.wg.Add(1)
	go s.taskProcessor()

	return s, nil
}

func (s *Server) CreateContainer(ctx context.Context, req *pb.CreateContainerRequest) (*pb.Container, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.containers) >= s.config.Resources.MaxContainers {
		return nil, status.Errorf(codes.ResourceExhausted, "maximum container limit reached")
	}

	containerID := fmt.Sprintf("obsidian-%s", uuid.New().String()[:8])
	
	dockerConfig := &docker.ContainerConfig{
		Image:       req.ImageId,
		Cmd:         req.Command,
		Env:         s.buildEnvVars(req.Environment),
		CPULimit:    s.parseCPULimit(req.Resources.CpuLimit),
		MemoryLimit: s.parseMemoryLimit(req.Resources.MemoryLimit),
		DiskLimit:   s.parseDiskLimit(req.Resources.DiskLimit),
		NetworkMode: req.Network,
		Labels: map[string]string{
			"obsidian.managed": "true",
			"obsidian.id":      containerID,
		},
		AutoRemove: false,
	}

	_, err := s.dockerClient.CreateContainer(ctx, containerID, dockerConfig)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create container: %v", err)
	}

	container := &Container{
		ID:        containerID,
		ImageID:   req.ImageId,
		State:     pb.ContainerState_CONTAINER_STATE_CREATED,
		Resources: req.Resources,
		CreatedAt: time.Now(),
		Labels:    dockerConfig.Labels,
	}

	s.containers[containerID] = container

	return s.containerToProto(container), nil
}

func (s *Server) StartContainer(ctx context.Context, req *pb.ContainerID) (*pb.Container, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	container, ok := s.containers[req.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "container not found: %s", req.Id)
	}

	if container.State != pb.ContainerState_CONTAINER_STATE_CREATED &&
		container.State != pb.ContainerState_CONTAINER_STATE_STOPPED {
		return nil, status.Errorf(codes.FailedPrecondition, "container is not in a startable state")
	}

	if err := s.dockerClient.StartContainer(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start container: %v", err)
	}

	container.State = pb.ContainerState_CONTAINER_STATE_RUNNING
	container.StartedAt = time.Now()

	return s.containerToProto(container), nil
}

func (s *Server) StopContainer(ctx context.Context, req *pb.ContainerID) (*pb.Container, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	container, ok := s.containers[req.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "container not found: %s", req.Id)
	}

	if container.State != pb.ContainerState_CONTAINER_STATE_RUNNING {
		return nil, status.Errorf(codes.FailedPrecondition, "container is not running")
	}

	if err := s.dockerClient.StopContainer(ctx, req.Id, 30*time.Second); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to stop container: %v", err)
	}

	container.State = pb.ContainerState_CONTAINER_STATE_STOPPED

	return s.containerToProto(container), nil
}

func (s *Server) RemoveContainer(ctx context.Context, req *pb.ContainerID) (*emptypb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	container, ok := s.containers[req.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "container not found: %s", req.Id)
	}

	if container.State == pb.ContainerState_CONTAINER_STATE_RUNNING {
		return nil, status.Errorf(codes.FailedPrecondition, "cannot remove running container")
	}

	if err := s.dockerClient.RemoveContainer(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove container: %v", err)
	}

	delete(s.containers, req.Id)

	return &emptypb.Empty{}, nil
}

func (s *Server) SubmitTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	container, ok := s.containers[req.ContainerId]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "container not found: %s", req.ContainerId)
	}

	if container.State != pb.ContainerState_CONTAINER_STATE_RUNNING {
		return nil, status.Errorf(codes.FailedPrecondition, "container is not running")
	}

	taskID := fmt.Sprintf("task-%s", uuid.New().String())
	timeout := s.config.Queue.TaskTimeout
	if req.Timeout != nil {
		timeout = req.Timeout.AsDuration()
	}

	task := &Task{
		ID:          taskID,
		ContainerID: req.ContainerId,
		Payload:     req.Payload,
		State:       pb.TaskState_TASK_STATE_QUEUED,
		CreatedAt:   time.Now(),
		Timeout:     timeout,
	}

	s.tasks[taskID] = task

	select {
	case s.taskQueue <- task:
	default:
		delete(s.tasks, taskID)
		return nil, status.Errorf(codes.ResourceExhausted, "task queue is full")
	}

	return &pb.TaskResponse{
		TaskId:      taskID,
		Result:      nil,
		Metrics:     nil,
		CompletedAt: nil,
	}, nil
}

func (s *Server) GetTaskStatus(ctx context.Context, req *pb.TaskID) (*pb.TaskStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[req.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "task not found: %s", req.Id)
	}

	return s.taskToStatus(task), nil
}

func (s *Server) CancelTask(ctx context.Context, req *pb.TaskID) (*emptypb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[req.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "task not found: %s", req.Id)
	}

	if task.State != pb.TaskState_TASK_STATE_QUEUED &&
		task.State != pb.TaskState_TASK_STATE_RUNNING {
		return nil, status.Errorf(codes.FailedPrecondition, "task cannot be cancelled in current state")
	}

	if task.Cancel != nil {
		task.Cancel()
	}

	task.State = pb.TaskState_TASK_STATE_CANCELLED
	task.CompletedAt = time.Now()

	return &emptypb.Empty{}, nil
}

func (s *Server) GetHealth(ctx context.Context, _ *emptypb.Empty) (*pb.HealthStatus, error) {
	healthy := true
	message := "healthy"

	if err := s.dockerClient.HealthCheck(ctx); err != nil {
		healthy = false
		message = fmt.Sprintf("docker unhealthy: %v", err)
	}

	return &pb.HealthStatus{
		Healthy:   healthy,
		Message:   message,
		LastCheck: timestamppb.Now(),
	}, nil
}

func (s *Server) GetMetrics(ctx context.Context, _ *emptypb.Empty) (*pb.Metrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	runningContainers := 0
	for _, c := range s.containers {
		if c.State == pb.ContainerState_CONTAINER_STATE_RUNNING {
			runningContainers++
		}
	}

	return &pb.Metrics{
		RunningContainers: int32(runningContainers),
		TotalContainers:   int32(len(s.containers)),
		QueueDepth:        int32(len(s.taskQueue)),
		CpuUtilization:    0.0,
		MemoryUtilization: 0.0,
	}, nil
}

func (s *Server) taskProcessor() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case task := <-s.taskQueue:
			s.processTask(task)
		}
	}
}

func (s *Server) processTask(task *Task) {
	s.mu.Lock()
	task.State = pb.TaskState_TASK_STATE_RUNNING
	task.StartedAt = time.Now()
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(s.ctx, task.Timeout)
	defer cancel()

	task.Cancel = cancel

	startTime := time.Now()

	result, err := s.executeTask(ctx, task)
	
	s.mu.Lock()
	defer s.mu.Unlock()

	task.CompletedAt = time.Now()
	task.Metrics = &pb.TaskMetrics{
		ExecutionTime: durationpb.New(time.Since(startTime)),
	}

	if err != nil {
		task.State = pb.TaskState_TASK_STATE_FAILED
		task.Error = err.Error()
	} else {
		task.State = pb.TaskState_TASK_STATE_SUCCESS
		task.Result = result
	}
}

func (s *Server) executeTask(ctx context.Context, task *Task) ([]byte, error) {
	return []byte("task completed successfully"), nil
}

func (s *Server) containerToProto(c *Container) *pb.Container {
	return &pb.Container{
		Id:        c.ID,
		ImageId:   c.ImageID,
		State:     c.State,
		Resources: c.Resources,
		CreatedAt: timestamppb.New(c.CreatedAt),
		StartedAt: timestamppb.New(c.StartedAt),
		Labels:    c.Labels,
	}
}

func (s *Server) taskToStatus(t *Task) *pb.TaskStatus {
	status := &pb.TaskStatus{
		TaskId:       t.ID,
		State:        t.State,
		ErrorMessage: t.Error,
		Metrics:      t.Metrics,
		CreatedAt:    timestamppb.New(t.CreatedAt),
	}

	if !t.StartedAt.IsZero() {
		status.StartedAt = timestamppb.New(t.StartedAt)
	}

	if !t.CompletedAt.IsZero() {
		status.CompletedAt = timestamppb.New(t.CompletedAt)
	}

	return status
}

func (s *Server) buildEnvVars(env map[string]string) []string {
	result := make([]string, 0, len(env)+len(s.config.Container.Env))
	
	for _, e := range s.config.Container.Env {
		result = append(result, fmt.Sprintf("%s=%s", e.Name, e.Value))
	}
	
	for k, v := range env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	
	return result
}

func (s *Server) parseCPULimit(limit string) int64 {
	return 1
}

func (s *Server) parseMemoryLimit(limit string) int64 {
	return 1073741824
}

func (s *Server) parseDiskLimit(limit string) int64 {
	return 10737418240
}

func (s *Server) Shutdown() {
	s.cancel()
	s.wg.Wait()
	s.dockerClient.Close()
}