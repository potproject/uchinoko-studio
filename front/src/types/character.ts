export type CharacterConfig = {
    general: {
        id: string;
        name: string;
        image: string;
    };
    multiVoice: boolean;
    voice: {
        type: string;
        identification: string;
        modelId: string;
        modelFile: string;
        speakerId: string;
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