import RecordRTC from "recordrtc";
import type { GeneralConfig } from "../types/general";
import type { RecordingContentInterface } from "./RecordingContentInterface";

export class RecordingContext implements RecordingContentInterface {
    private stream: MediaStream;
    private mediaRecorder!: RecordRTC;
    private mimeType: string;

    private threshold: number;
    private silentThreshold: number;
    private audioMinLength: number;

    private audioContext: AudioContext;
    private sampleRate: number;
    private isRecordingAllow: boolean = true;
    private recordStartTime: number = 0;
    private recordStopTime: number = 0;

    public onSpeakingStart: () => void = () => { };
    public onSpeakingEnd: (ignore: boolean) => void = () => { };
    public onDataAvailable: (event: BlobEvent) => void = () => { };
    public onText: (text: string) => void = () => { };

    constructor(stream: MediaStream, mimeType: string, generalConfig: GeneralConfig) {
        this.stream = stream;
        /** @ts-ignore */
        this.audioContext = new (window.AudioContext || window.webkitAudioContext)();
        this.sampleRate = this.stream.getAudioTracks()[0].getSettings().sampleRate ?? 44100;
        this.mimeType = mimeType;
        this.threshold = generalConfig.transcription.autoSetting.threshold ?? 0.02;
        this.silentThreshold = generalConfig.transcription.autoSetting.silentThreshold ?? 1;
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
        this.isRecordingAllow = check;
        const state = this.mediaRecorder.getState();
        if (state === 'recording') {
            this.mediaRecorder.stopRecording();
            this.onSpeakingEnd(true);
            return;
        }
    }

    async init() {
        await this.audioContext.audioWorklet.addModule('audio-worklet-processors.js');
        const volumeNode = new AudioWorkletNode(this.audioContext, 'volume-processor', {
            processorOptions: {
                sampleRate: this.sampleRate,
                threshold: this.threshold, // 音量の閾値
                silentThreshold: this.silentThreshold, // 無音状態の閾値
            }
        });
        const source = this.audioContext.createMediaStreamSource(this.stream);
        volumeNode.port.onmessage = event => {
            const speak = event.data.speak;
            if (speak && this.isRecordingAllow) {
                this.mediaRecorder.startRecording();
                this.recordStartTime = Date.now();
                this.onSpeakingStart();
            } else {
                const state = this.mediaRecorder.getState();
                if (state === 'recording') {
                    this.recordStopTime = Date.now();
                    this.mediaRecorder.stopRecording(() => {
                        const blob = this.mediaRecorder.getBlob();
                        const event = new BlobEvent('dataavailable', { data: blob });
                        const recordingTime = this.recordStopTime - this.recordStartTime;
                        if (event.data.size > 0 && recordingTime >= (this.audioMinLength*1000)) {
                            this.onDataAvailable(event);
                        }
                        this.create();
                    });
                    const recordingTime = this.recordStopTime - this.recordStartTime;
                    if (recordingTime < (this.audioMinLength*1000)) {
                        this.onSpeakingEnd(true);
                        return;
                    }
                    this.onSpeakingEnd(false);
                }
            }
        };
        source.connect(volumeNode).connect(this.audioContext.destination);
    }
}