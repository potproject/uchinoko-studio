import type { GeneralConfig } from "../types/general";
import type { RecordingContentInterface } from "./RecordingContentInterface";

export class RecognitionContent implements RecordingContentInterface {
    private recognition: any;

    private isRecordingAllow: boolean = true;
    private isRecognitionRunning: boolean = false;
    private result: boolean = false;
    private isStart: boolean = false;

    public onSpeakingStart: () => void = () => { };
    public onSpeakingEnd: (ignore: boolean) => void = () => { };
    public onDataAvailable: (event: BlobEvent) => void = () => { };
    public onText: (text: string) => void = () => { };

    constructor(stream: MediaStream, mimeType: string, generalConfig: GeneralConfig) {
        void stream;
        void mimeType;

        /** @ts-ignore */
        this.recognition = new (window.SpeechRecognition || window.webkitSpeechRecognition)();
        this.recognition.lang = generalConfig.language;
        this.recognition.interimResults = false;
        this.recognition.continuous = false;
        this.recognition.onresult = (event: any) => {
            if (!this.isRecordingAllow) {
                this.result = false;
                this.isStart = false;
                return;
            }
            if (!this.isStart) {
                this.onSpeakingStart();
            }
            this.onSpeakingEnd(false);
            const text = event.results[0][0].transcript;
            this.result = true;
            this.isStart = false;
            this.onText(text);
        }

        this.recognition.onspeechstart = () => {
            if (!this.isRecordingAllow || this.isStart) {
                return;
            }
            this.onSpeakingStart();
            this.isStart = true;
        }

        this.recognition.onspeechend = () => {
        }

        this.recognition.onend = () => {
            this.isRecognitionRunning = false;
            if (this.result === false && this.isStart === true) {
                this.onSpeakingEnd(true);
            }
            const shouldRestart = this.result === false && this.isRecordingAllow;
            this.isStart = false;
            this.result = false;
            if (shouldRestart) {
                this.startRecognition();
            }
        }

        this.recognition.onerror = (event: any) => {
            console.log("error", event);
            this.isRecognitionRunning = false;
        }
        this.startRecognition();
    }

    private startRecognition() {
        if (!this.isRecordingAllow || this.isRecognitionRunning) {
            return;
        }
        this.recognition.start();
        this.isRecognitionRunning = true;
    }

    public changeRecordingAllow(check: boolean) {
        this.isRecordingAllow = check;
        if (check) {
            this.startRecognition();
            return;
        }
        if (this.isRecognitionRunning) {
                this.recognition.stop();
                this.isRecognitionRunning = false;
        }
    }

    async init() {
    }

    dispose() {
        this.isRecordingAllow = false;
        if (this.isRecognitionRunning) {
            this.recognition.stop();
            this.isRecognitionRunning = false;
        }
        this.recognition.onresult = null;
        this.recognition.onspeechstart = null;
        this.recognition.onspeechend = null;
        this.recognition.onend = null;
        this.recognition.onerror = null;
    }
}
