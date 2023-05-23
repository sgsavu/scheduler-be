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
    BatchSize: number,
    CreationTime: string,
    Epochs: number,
    ModelName: string,
    SampleRate: number,
    Status: TaskStatus,
    TerminationTime?: string
    Type: TaskType
    result?: Blob
}