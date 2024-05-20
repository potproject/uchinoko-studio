export type Message = {
    type: 'my' | 'your' | 'error' | 'my-img';
    voiceIndex: number|null;
    text: string;
    img?: string;
    loading: boolean;
    speaking: boolean;
    chunk: boolean;
}

export type ChunkMessage = {
    type: 'change-character' | 'change-behavior' | 'chat';
    text : string;
}