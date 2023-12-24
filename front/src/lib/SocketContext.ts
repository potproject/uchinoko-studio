type TextMessage = {
    type: 'connection' | 'chat-response' | 'chat-request' | 'chat-response-chunk' | 'finish' | 'error';
    text: string;
};

export class SocketContext{
    private socket: WebSocket;
    
    public onConnected: () => void = () => {};
    public onBinary: (data: ArrayBuffer) => void = () => {};
    public onText: (data: TextMessage) => void = () => {};
    public onClosed: () => void = () => {};

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
            } else if (data.type === 'chat-request') {
                console.log('chat-request', data.text);
            } else if (data.type === 'chat-response-chunk') {
                console.log('chat-response-chunk', data.text);
            } else if (data.type === 'finish') {
                console.log('finish');
            } else if (data.type === 'error') {
                console.log('error', data.text);
            }
            this.onText(data);
        }

        this.socket.onclose = () => {
            this.onClosed();
        }
    }

    public sendBinary(data: any){
        this.socket.send(data);
    }
}  