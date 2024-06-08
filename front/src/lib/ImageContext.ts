import { ImageResize } from "./ImageResize";

export class ImageContext {
    public onLoadStart: (file: File) => void = (file: File) => {};
    public onLoadEnd: (arrayBuffer: ArrayBuffer) => void = (arrayBuffer: ArrayBuffer) => {};
    async upload() {
        const input = globalThis.document.createElement("input");
        input.type = "file";
        input.accept = "image/jpeg, image/png";
        input.onchange = async () => {
            if (!input.files || input.files.length === 0) {
                return;
            }
            const file = input.files[0];
            const reader = new FileReader();
            reader.onload = async () => {
                const arrayBuffer = reader.result as ArrayBuffer;
                this.onLoadStart(file);
                const resizeArrayBuffer = await ImageResize.run(arrayBuffer);
                this.onLoadEnd(resizeArrayBuffer);
            };
            reader.readAsArrayBuffer(file);
        };
        input.click();
    }
}