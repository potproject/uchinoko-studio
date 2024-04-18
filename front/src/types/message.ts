export type Message = {
    type: 'my' | 'your' | 'error';
    voiceIndex: number|null;
    text: string;
    loading: boolean;
    speaking: boolean;
    chunk: boolean;
}