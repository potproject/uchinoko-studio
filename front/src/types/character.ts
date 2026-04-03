export type CharacterConfig = {
    general: {
        id: string;
        name: string;
    };
    multiVoice: boolean;
    voice: {
        name: string;
        type: "voicevox"| "bertvits2" | "irodori-tts" | "stylebertvits2" | "nijivoice" | "google-text-to-speech" | "openai-speech";
        image: string;
        identification: string;
        modelId: string;
        modelFile: string;
        speakerId: string;
        referenceAudioPath: string;
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
        temperature: {
            enable: boolean;
            value: number;
        }
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
    memory: {
        enabled: boolean;
        maxItemsInPrompt: number;
        enableRelationshipMemory: boolean;
        enableSessionSummary: boolean;
        enableSemanticSearch: boolean;
        embeddingModel: string;
        allowSensitiveMemory: boolean;
    };
};

export type CharacterConfigList = {
    characters: CharacterConfig[];
};

export type MemoryItem = {
    id: string;
    characterId: string;
    ownerId: string;
    scope: "character" | "relationship";
    kind: string;
    content: string;
    keywordsText: string;
    pinned: boolean;
    confidence: number;
    salience: number;
    source: string;
    updatedAt: string;
};

export type MemoryItemList = {
    items: MemoryItem[];
};

export type SessionSummary = {
    ownerId: string;
    characterId: string;
    sessionId: string;
    summary: string;
    updatedAt: string;
};
