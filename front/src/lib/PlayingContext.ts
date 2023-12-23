const bufferSeconds = 0.1;

export class PlayingContext {
    private audioContext: AudioContext;
    private nextTime = 0;
    private playing = false;
    private sendFinish = false;
    private latestAudioBufferSourceNode: AudioBufferSourceNode|null = null;

    public onSpeakingStart: () => void = () => {};
    public onSpeakingEnd: () => void = () => {};

    constructor(AudioContext: AudioContext) {
        this.audioContext = AudioContext;
    }

    getAudioContext() {
        return this.audioContext;
    }

    // sendFinishAction これが呼ばれると再生が終了したときに playingRefresh が呼ばれる
    sendFinishAction() {
        this.sendFinish = true;
    }

    isPlaying() {
        return this.playing;
    }

    private playingRefresh(){
        this.playing = false;
        this.nextTime = 0;
        this.sendFinish = false;
        this.onSpeakingEnd();
    }

    playWAV(arrayBuffer: ArrayBuffer) {
        if (!this.playing) {
            this.onSpeakingStart();
            this.playing = true;
        }
        this.audioContext.decodeAudioData(arrayBuffer, audioBuffer => {
            const source = this.audioContext.createBufferSource();
            source.buffer = audioBuffer;
            source.onended = (event) => {
                if (event.target === this.latestAudioBufferSourceNode && this.sendFinish) {
                    this.playingRefresh();
                }
            };
            source.connect(this.audioContext.destination);

            // シームレスな再生のために次の開始時間を設定
            if (this.nextTime === 0) {
                this.nextTime = this.audioContext.currentTime;
            }

            source.start(this.nextTime);
            this.nextTime += (audioBuffer.duration + bufferSeconds);

            this.latestAudioBufferSourceNode = source;
        });
    }
}
