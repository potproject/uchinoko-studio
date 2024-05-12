export type GeneralConfig = {
    transcription: {
        type: string;
        method: "auto"|"pushToTalk"
        autoSetting: {
            threshold: number;
            silentThreshold: number;
            audioMinLength: number;
        };
    };
};