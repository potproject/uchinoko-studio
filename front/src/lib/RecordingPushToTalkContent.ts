import RecordRTC from "recordrtc";
// 1.3秒以上の音声データを送信
const audioMinLength = 1300;

export class RecordingPushToTalkContext {
    private stream: MediaStream;
    private mimeType: string;
    private mediaRecorder!: RecordRTC;
    private recordStartTime: number = 0;
    private recordStopTime: number = 0;

    public onSpeakingStart: () => void = () => { };
    public onSpeakingEnd: (ignore: boolean) => void = () => { };
    public onDataAvailable: (event: BlobEvent) => void = () => { };

    constructor(stream: MediaStream, mimeType: string) {
        this.stream = stream;
        this.mimeType = mimeType;
        /** @ts-ignore */
        this.audioContext = new (window.AudioContext || window.webkitAudioContext)();
        this.create();
    }

    private create() {
        this.mediaRecorder = new RecordRTC(this.stream, {
            type: 'audio',
            mimeType: this.mimeType as any,
            recorderType: RecordRTC.StereoAudioRecorder,
            numberOfAudioChannels: 1,
            disableLogs: true,
        });
    }

    public changeRecordingAllow(check: boolean) {
        const state = this.mediaRecorder.getState();
        if (check && (state === 'inactive' || state === 'stopped')) {
            this.mediaRecorder.startRecording();
            this.recordStartTime = Date.now();
            this.onSpeakingStart();
        } else {
            if (state === 'recording') {
                this.mediaRecorder.stopRecording(() => {
                    const blob = this.mediaRecorder.getBlob();
                    const event = new BlobEvent('dataavailable', { data: blob });
                    this.onDataAvailable(event);
                    this.create();
                });
                this.recordStopTime = Date.now();
                const recordingTime = this.recordStopTime - this.recordStartTime;
                if (recordingTime >= audioMinLength) {
                    this.onSpeakingEnd(false);
                } else {
                    this.onSpeakingEnd(true);
                }
            }
        }
    }

    async init() {

    }
}