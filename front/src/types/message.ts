export type Message = {
    type: 'my' | 'your' | 'error';
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

export const MessageConstants = {
    empty: '',
    uploadImage: '画像をアップロード中です...',
    disconnected: '接続が切断されました。続けるにはページをリロードしてください。',
    speakingStart: '話し中...',
    speakingEnd: '音声認識中...',

}