package avsExecutionManager

/*
func (em *AvsExecutionManager) Start(ctx context.context) error {
	em.logger.Sugar().Infow("Starting AvsExecutionManager",
		"secureConnection", em.config.SecureConnection,
		"refreshInterval", em.config.PeerRefreshInterval,
	)

	err := em.rpcServer.Start(ctx)
	if err != nil {
		return err
	}

	go em.processTaskQueue(ctx)
	go em.refreshExecutorClientsLoop(ctx)

	return nil
}

func (em *AvsExecutionManager) refreshExecutorClientsLoop(ctx context.context) {
	ticker := time.NewTicker(em.config.PeerRefreshInterval)
	defer ticker.Stop()

	sugar := em.logger.Sugar()
	sugar.Info("Starting executor client refresh loop")

	em.refreshExecutorClients()

	for {
		select {
		case <-ctx.Done():
			sugar.Info("Stopping executor client refresh loop")
			return
		case <-ticker.C:
			em.refreshExecutorClients()
		}
	}
}

func (em *AvsExecutionManager) processTaskQueue(ctx context.context) {
	sugar := em.logger.Sugar()
	sugar.Info("Starting task processing loop")

	for {
		select {
		case <-ctx.Done():
			sugar.Info("Stopping task processing loop")
			return
		case task, ok := <-em.taskQueue:
			if !ok {
				sugar.Warn("task queue channel closed, exiting")
				return
			}

			go em.processTask(ctx, task)
		}
	}
}

func (em *AvsExecutionManager) processTask(ctx context.context, task *types.task) {
	sugar := em.logger.Sugar()
	sugar.Infow("Processing task", "taskId", task.TaskId)
	em.running.Store(task.TaskId, task)

	sig, err := em.signer.SignMessage(task.Payload)
	if err != nil {
		sugar.Errorw("Failed to sign task payload",
			zap.String("taskId", task.TaskId),
			zap.Error(err),
		)
		return
	}

	aggregatorUrl := fmt.Sprintf("localhost:%d", em.rpcServer.RpcConfig.GrpcPort)
	if em.config.AggregatorUrl != "" {
		sugar.Infow("Using custom aggregator URL",
			zap.String("aggregatorUrl", em.config.AggregatorUrl),
		)
		aggregatorUrl = em.config.AggregatorUrl
	}

	taskSubmission := &executorV1.TaskSubmission{
		TaskId:            task.TaskId,
		AvsAddress:        task.AVSAddress,
		AggregatorAddress: em.config.AggregatorOperatorAddress,
		Payload:           task.Payload,
		AggregatorUrl:     aggregatorUrl,
		Signature:         sig,
	}

	var wg sync.WaitGroup
	for addr, execClient := range em.execClients {
		wg.Add(1)

		go func(address string, client executorV1.ExecutorServiceClient, wg *sync.WaitGroup) {
			defer wg.Done()
			fmt.Printf("Submitting task: %+v\n", taskSubmission)
			res, err := client.SubmitTask(ctx, taskSubmission)
			if err != nil {
				sugar.Errorw("Failed to submit task to executor",
					zap.String("executorAddress", address),
					zap.String("taskId", task.TaskId),
					zap.Error(err),
				)
				return
			}
			if !res.Success {
				sugar.Errorw("task submission failed",
					zap.String("executorAddress", address),
					zap.String("taskId", task.TaskId),
					zap.String("message", res.Message),
				)
				return
			}
			sugar.Debugw("Successfully submitted task to executor",
				zap.String("executorAddress", address),
				zap.String("taskId", task.TaskId),
			)

		}(addr, execClient, &wg)
	}
	wg.Wait()
	sugar.Infow("task submission completed", zap.String("taskId", task.TaskId))
}

func (em *AvsExecutionManager) SubmitTaskResult(
	ctx context.context,
	result *aggregatorV1.TaskResult,
) (*v1.SubmitAck, error) {
	sugar := em.logger.Sugar()
	taskID := result.TaskId

	value, ok := em.running.Load(taskID)
	if !ok {
		sugar.Warnw("Received result for unknown task", "task_id", taskID)
		return &v1.SubmitAck{Success: false, Message: "unknown task"}, nil
	}

	em.running.Delete(taskID)
	task := value.(*types.task)

	taskResult := &types.TaskResult{
		TaskId:        task.TaskId,
		AvsAddress:    task.AVSAddress,
		CallbackAddr:  task.CallbackAddr,
		OperatorSetId: task.OperatorSetId,
		Output:        result.Output,
		ChainId:       task.ChainId,
		BlockNumber:   task.BlockNumber,
		BlockHash:     task.BlockHash,
	}

	select {
	case em.resultQueue <- taskResult:
		sugar.Infow("task result accepted", "task_id", taskID)
		return &v1.SubmitAck{Success: true, Message: "ok"}, nil
	case <-time.After(1 * time.Second):
		sugar.Errorw("Failed to enqueue task result (channel full or closed)", "task_id", taskID)
		return &v1.SubmitAck{Success: false, Message: "enqueue error"}, nil
	case <-ctx.Done():
		sugar.Warnw("context cancelled while enqueueing result", "task_id", taskID)
		return &v1.SubmitAck{Success: false, Message: "context cancelled"}, nil
	}
}

func (em *AvsExecutionManager) refreshExecutorClients() {
	sugar := em.logger.Sugar()

	peers, err := em.peeringDataFetcher.ListExecutorOperators()
	if err != nil {
		sugar.Errorw("Failed to list executor peers", "error", err)
		return
	}

	newClientCount := 0

	for _, peer := range peers {
		if _, exists := em.execClients[peer.PublicKey]; !exists {
			addr := fmt.Sprintf("%s:%d", peer.NetworkAddress, peer.Port)

			// TODO - SecureConnection should always be used unless the address contains 'localhost' or '127.0.0.1'
			client, err := executorClient.NewExecutorClient(addr, !em.config.SecureConnection)
			if err != nil {
				// TODO: emit metric
				sugar.Errorw("Failed to create executor client",
					zap.String("address", addr),
					zap.String("publicKey", peer.PublicKey),
					zap.String("operatorAddress", peer.OperatorAddress),
				)
				continue
			}
			em.execClients[peer.PublicKey] = client
			newClientCount++
			// TODO: emit metric
			sugar.Infow("Registered new executor client", "public_key", peer.PublicKey)
		}
	}

	if newClientCount > 0 {
		sugar.Infow("Refreshed executor clients", "newClients", newClientCount, "totalClients", len(em.execClients))
	}
}*/
