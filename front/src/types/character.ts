export type CharacterConfig = {
    general: {
        id: string;
        name: string;
    };
    multiVoice: boolean;
    voice: {
        type: string;
        image: string;
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
    history: string;
};

export type CharacterConfigList = {
    characters: CharacterConfig[];
};