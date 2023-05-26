export enum TaskType {
    TRAINING = 'TRAINING',
    INFERRING = 'INFERRING'
}

export enum TaskStatus {
    WORKING = 'WORKING',
    DONE = 'DONE',
    FAILED = 'FAILED'
}

export interface Task {
    ID: string,
    BatchSize?: number,
    CreationTime: string,
    ExpiryTime?: string,
    FailureReason?: string,
    Name: string,
    Status: TaskStatus,
    TerminationTime?: string
    Type: TaskType
    result?: Blob
}