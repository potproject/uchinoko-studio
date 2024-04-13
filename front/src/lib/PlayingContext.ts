const bufferSeconds = 0.1;

export class PlayingContext {
    private audioContext: AudioContext;
    private nextTime = 0;
    private playing = false;
    private sendFinish = false;
    private latestAudioBufferSourceNode: AudioBufferSourceNode|null = null;

    public onSpeakingStart: () => void = () => {};
    public onSpeakingEnd: () => void = () => {};
    public onSpeackingChunkStart: () => void = () => {};

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

            // 遅延対策
            // すでに次の開始時間が過ぎていたら、次の開始時間を現在時刻にする
            if (this.nextTime < this.audioContext.currentTime) {
                this.nextTime = this.audioContext.currentTime;
            }

            source.start(this.nextTime);

            setTimeout(() => {
                this.onSpeackingChunkStart();
            }, (this.nextTime - this.audioContext.currentTime) * 1000);
            this.nextTime += (audioBuffer.duration + bufferSeconds);
            this.latestAudioBufferSourceNode = source;
        });
    }

    playPCM(arrayBuffer: ArrayBuffer) {
        // 16ビットのサンプルをFloat32Arrayに変換
        const int16Array = new Int16Array(arrayBuffer);
        const float32Array = new Float32Array(int16Array.length);

        for (let i = 0; i < int16Array.length; i++) {
            // Int16の範囲(-32768 to 32767)をFloat32の範囲(-1.0 to 1.0)に正規化
            float32Array[i] = int16Array[i] / 32768;
        }

        // AudioBufferの作成
        const audioBuffer = this.audioContext.createBuffer(1, float32Array.length, 22050);
        audioBuffer.getChannelData(0).set(float32Array);

        // AudioBufferSourceNodeの作成と設定
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
        // 次のオーディオチャンクの再生開始時間を更新
        this.nextTime += audioBuffer.duration;

        this.latestAudioBufferSourceNode = source;
    }
}
