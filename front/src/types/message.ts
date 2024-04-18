export type Message = {
    type: 'my' | 'your' | 'error';
    voiceIndex: number|null;
    text: string;
    textChunk: string[];
    loading: boolean;
    speaking: boolean;
    chunk: boolean;
}