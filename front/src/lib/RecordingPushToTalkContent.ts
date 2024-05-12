import RecordRTC from "recordrtc";
import type { GeneralConfig } from "../types/general";

export class RecordingPushToTalkContext {
    private stream: MediaStream;
    private mimeType: string;
    private mediaRecorder!: RecordRTC;
    private recordStartTime: number = 0;
    private recordStopTime: number = 0;

    private audioMinLength: number;

    public onSpeakingStart: () => void = () => { };
    public onSpeakingEnd: (ignore: boolean) => void = () => { };
    public onDataAvailable: (event: BlobEvent) => void = () => { };

    constructor(stream: MediaStream, mimeType: string, generalConfig: GeneralConfig) {
        this.stream = stream;
        this.mimeType = mimeType;
        /** @ts-ignore */
        this.audioContext = new (window.AudioContext || window.webkitAudioContext)();
        this.audioMinLength = generalConfig.transcription.autoSetting.audioMinLength ?? 1.3;
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
                if (recordingTime >= (this.audioMinLength*1000)) {
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