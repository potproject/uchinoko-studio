// 1.3秒以上の音声データを送信
const audioMinLength = 1300;

export class RecordingContext{
    private stream: MediaStream;
    private mediaRecorder: MediaRecorder;
    private audioContext: AudioContext;
    private sampleRate: number;
    private isRecordingAllow: boolean = true;
    private recordStartTime: number = 0;
    private recordStopTime: number = 0;

    public onSpeakingStart: () => void = () => {};
    public onSpeakingEnd: (ignore: boolean) => void = () => {};
    public onDataAvailable: (event: BlobEvent) => void = () => {};

    constructor(stream: MediaStream, mimeType: string){
        this.stream = stream;
        /** @ts-ignore */
        this.audioContext = new (window.AudioContext || window.webkitAudioContext)();
        this.sampleRate = this.stream.getAudioTracks()[0].getSettings().sampleRate ?? 44100;
        this.mediaRecorder = new MediaRecorder(this.stream, { mimeType });
    }

    public changeRecordingAllow(check: boolean){
        this.isRecordingAllow = check;
    }

    async init(){
        await this.audioContext.audioWorklet.addModule('audio-worklet-processors.js');
        const volumeNode = new AudioWorkletNode(this.audioContext, 'volume-processor', {
            processorOptions: {
                sampleRate: this.sampleRate,
                threshold: 0.02,
            }
        });
        const source = this.audioContext.createMediaStreamSource(this.stream);
        volumeNode.port.onmessage = event => {
            const speak = event.data.speak;
            if (speak && this.isRecordingAllow) {
                this.mediaRecorder.start();
                this.recordStartTime = Date.now();
                this.onSpeakingStart();
            }else{
                if (this.mediaRecorder.state !== 'inactive'){
                    this.recordStopTime = Date.now();
                    this.mediaRecorder.stop();
                    const recordingTime = this.recordStopTime - this.recordStartTime;
                    if (recordingTime < audioMinLength) {
                        this.onSpeakingEnd(true);
                        return;
                    }
                    this.onSpeakingEnd(false);
                }
            }
        };
        source.connect(volumeNode).connect(this.audioContext.destination);
        this.mediaRecorder.ondataavailable = event => {
            const recordingTime = this.recordStopTime - this.recordStartTime;
            if (event.data.size > 0 && recordingTime >= audioMinLength) {
                this.onDataAvailable(event);
            }
        };
    }
}