/**
 * Example Lambda function for AVS performer
 * This demonstrates how an AVS can implement custom logic
 */

exports.handler = async (event, context) => {
    console.log('Received event:', JSON.stringify(event, null, 2));
    
    try {
        // Extract AVS-specific information
        const { avsAddress, taskId, payload } = event;
        
        // Parse the payload (if it's a string)
        let taskData;
        if (typeof payload === 'string') {
            try {
                taskData = JSON.parse(payload);
            } catch (e) {
                taskData = { raw: payload };
            }
        } else {
            taskData = payload || {};
        }
        
        // Example: Perform some computation based on the task
        const result = await processTask(taskData);
        
        // Return successful response
        return {
            statusCode: 200,
            body: JSON.stringify({
                success: true,
                avsAddress,
                taskId,
                result,
                processedAt: new Date().toISOString(),
                lambdaRequestId: context.awsRequestId
            })
        };
        
    } catch (error) {
        console.error('Error processing task:', error);
        
        return {
            statusCode: 500,
            body: JSON.stringify({
                success: false,
                error: error.message,
                lambdaRequestId: context.awsRequestId
            })
        };
    }
};

/**
 * Example task processing function
 * Replace this with your AVS-specific logic
 */
async function processTask(taskData) {
    // Example 1: Simple computation
    if (taskData.operation === 'compute') {
        const { a, b } = taskData.params || {};
        return {
            sum: (a || 0) + (b || 0),
            product: (a || 0) * (b || 0),
            timestamp: Date.now()
        };
    }
    
    // Example 2: Data transformation
    if (taskData.operation === 'transform') {
        const { input } = taskData.params || {};
        return {
            original: input,
            uppercase: input?.toUpperCase(),
            length: input?.length || 0,
            reversed: input?.split('').reverse().join('')
        };
    }
    
    // Example 3: Async operation simulation
    if (taskData.operation === 'async') {
        // Simulate async work
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        return {
            message: 'Async operation completed',
            duration: '1000ms',
            random: Math.random()
        };
    }
    
    // Default response for unknown operations
    return {
        message: 'Task processed successfully',
        operation: taskData.operation || 'unknown',
        echo: taskData
    };
}