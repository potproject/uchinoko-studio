import type { GeneralConfig } from "../types/general";
import type { RecordingContentInterface } from "./RecordingContentInterface";

export class RecognitionContent implements RecordingContentInterface {
    private stream: MediaStream;
    private recognition: any;

    private state: Boolean = false;
    private result: Boolean = false;
    private isStart: Boolean = false;

    public onSpeakingStart: () => void = () => { };
    public onSpeakingEnd: (ignore: boolean) => void = () => { };
    public onDataAvailable: (event: BlobEvent) => void = () => { };
    public onText: (text: string) => void = () => { };

    constructor(stream: MediaStream, mimeType: string, generalConfig: GeneralConfig) {
        this.stream = stream;
        /** @ts-ignore */
        this.audioContext = new (window.AudioContext || window.webkitAudioContext)();

        /** @ts-ignore */
        this.recognition = new (window.SpeechRecognition || window.webkitSpeechRecognition)();
        this.recognition.lang = 'ja-JP';
        this.recognition.interimResults = false;
        this.recognition.continuous = false;
        this.recognition.onresult = (event: any) => {
            if(!this.isStart) {
                this.onSpeakingStart();
            }
            this.onSpeakingEnd(false);
            const text = event.results[0][0].transcript;
            this.result = true;
            this.isStart = false;
            this.onText(text);
        }

        this.recognition.onspeechstart = () => {
            this.onSpeakingStart();
            this.isStart = true;
        }

        this.recognition.onspeechend = () => {
        }

        this.recognition.onend = () => {
            if (this.result === false && this.isStart === true) {
                this.onSpeakingEnd(true);
            }
            if (this.result === false){
                this.recognition.start();
            }
            this.result = false;
        }

        this.recognition.onerror = (event: any) => {
            console.log("error", event);
            if(!this.state) {
                this.recognition.stop();
            }
            this.state = false;
        }
        this.state = true;
        this.recognition.start();
    }

    public changeRecordingAllow(check: boolean) {
        if (check && !this.state) {
            this.recognition.start();
            this.state = true;
        } else {
            if (this.state) {
                this.recognition.stop();
                this.state = false;
            }
        }
    }

    async init() {
    }
}