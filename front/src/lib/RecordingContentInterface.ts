export interface RecordingContentInterface {
    onSpeakingStart: () => void;
    onSpeakingEnd: (ignore: boolean) => void;
    onDataAvailable: (event: BlobEvent) => void;
    onText: (text: string) => void;

    changeRecordingAllow(check: boolean): void;
    init(): Promise<void>;
}