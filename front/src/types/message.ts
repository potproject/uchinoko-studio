export type Message = {
    type: 'my' | 'your' | 'error';
    voiceIndex: number|null;
    text: string;
    loading: boolean;
    speaking: boolean;
    chunk: boolean;
}

export type ChunkMessage = {
    type: 'change-character' | 'change-behavior' | 'chat';
    text : string;
}