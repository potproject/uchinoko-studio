export type CharacterConfig = {
    general: {
        id: string;
        name: string;
    };
    multiVoice: boolean;
    voice: {
        name: string;
        type: "voicevox"| "bertvits2" | "stylebertvits2" | "nijivoice" | "google-text-to-speech" | "openai-speech";
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
        maxHistory: number;
        limit:{
            day: {
                request: number;
                token: number;
            },
            hour: {
                request: number;
                token: number;
            },
            minute: {
                request: number;
                token: number;
            }
        }
    };
};

export type CharacterConfigList = {
    characters: CharacterConfig[];
};