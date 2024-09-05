export type GeneralConfig = {
    language: "ja-JP";
    soundEffect: boolean;
    characterOutputChange: boolean;
    transcription: {
        type: "openai_speech_to_text"|"google_speech_to_text"|"vosk_server"|"speech_recognition"
        method: "auto"|"pushToTalk"
        autoSetting: {
            threshold: number;
            silentThreshold: number;
            audioMinLength: number;
        };
    };
};