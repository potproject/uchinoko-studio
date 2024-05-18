export type CharacterConfig = {
    general: {
        id: string;
        name: string;
    };
    multiVoice: boolean;
    voice: {
        type: "voicevox"| "bertvits2" | "stylebertvits2" | "google-text-to-speech" | "openai-speech";
        image: string;
        identification: string;
        modelId: string;
        modelFile: string;
        speakerId: string;
        backgroundImagePath: string;
        behavior: {
            identification: string;
            imagePath: string;
        }[];
    }[];
    chat: {
        type: string;
        model: string;
        systemPrompt: string;
    };
};

export type CharacterConfigList = {
    characters: CharacterConfig[];
};