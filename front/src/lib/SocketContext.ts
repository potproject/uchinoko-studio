import type { CharacterConfig } from "../types/character";

type TextMessage = {
    type: 'connection' | 'chat-response' | 'chat-request' | 'chat-response-change-character' | 'chat-response-change-behavior' | 'chat-response-chunk' | 'finish' | 'error';
    text: string;
};

export class SocketContext{
    private socket: WebSocket;

    public mimeType: string = 'audio/wav';
    
    public onConnected: () => void = () => {};
    public onBinary: (data: ArrayBuffer) => void = () => {};
    public onClosed: () => void = () => {};

    public onChatRequest: (text: string) => void = () => {};
    public onChatResponse: (text: string) => void = () => {};
    public onChatResponseChangeCharacter: (text: string) => void = () => {};
    public onChatResponseChangeBehavior: (text: string) => void = () => {};
    public onChatResponseChunk: (text: string) => void = () => {};
    public onFinish: () => void = () => {};
    public onError: (text: string) => void = () => {};

    constructor(url: string){
        this.socket = new WebSocket(url);
        this.socket.binaryType = "arraybuffer";
        this.socket.onmessage = (event) => {
            // binaryの場合
            if (event.data instanceof ArrayBuffer) {
                this.onBinary(event.data);
                return;
            }
            // textの場合
            const data = JSON.parse(event.data) as TextMessage;
            if (data.type === 'connection') {
                this.onConnected();
                return;
            } else if (data.type === 'chat-response') {
                console.log('chat-response', data.text);
                this.onChatResponse(data.text);
            } else if (data.type === 'chat-request') {
                console.log('chat-request', data.text);
                this.onChatRequest(data.text);
            } else if (data.type === 'chat-response-change-character') {
                console.log('chat-response-change-character', data.text);
                this.onChatResponseChangeCharacter(data.text);
            } else if (data.type === 'chat-response-change-behavior') {
                console.log('chat-response-change-behavior', data.text);
                this.onChatResponseChangeBehavior(data.text);
            } else if (data.type === 'chat-response-chunk') {
                console.log('chat-response-chunk', data.text);
                this.onChatResponseChunk(data.text);
            } else if (data.type === 'finish') {
                console.log('finish', data.text);
                this.onFinish();
            } else if (data.type === 'error') {
                console.log('error', data.text);
                this.onError(data.text);
            }
        }

        this.socket.onclose = () => {
            this.onClosed();
        }
    }

    public static async connect(selectCharacter: CharacterConfig): Promise<SocketContext> {
        const wsTLS = location.protocol === 'https:' ? 'wss' : 'ws';
    
        // chromeの場合はcompressを有効にする
        const ua = window.navigator.userAgent.toLowerCase();
        const isChrome = ua.indexOf('chrome') != -1 && ua.indexOf('edge') == -1;
        const compressed = isChrome ? '/compressed' : '';

        const url = `${wsTLS}://${location.host}/v1/ws/talk/${selectCharacter.general.id}/${selectCharacter.general.id}${compressed}`;
        const socket = new SocketContext(url);
        await new Promise(resolve => {
            socket.onConnected = () => {
                resolve(socket);
            }
        });
        return socket;
    }

    public sendBinary(data: string | ArrayBufferLike | ArrayBufferView | Blob){
        this.socket.send(data);
    }

    public sendText(text: string){
        const data = JSON.stringify({text});
        this.socket.send(data);
    }
}  