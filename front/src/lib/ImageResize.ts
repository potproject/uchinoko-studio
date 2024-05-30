import { decode as jpgDecode, encode as jpgEncode } from '@jsquash/jpeg';
import { decode as pngDecode } from '@jsquash/png';
import resize from '@jsquash/resize';

export class ImageResize {
    private static isType(arrayBuffer: ArrayBuffer): "jpg" | "png" | "unknown" {
        const uint8Array = new Uint8Array(arrayBuffer);

        // Check for JPG (JPEG) file signature: FF D8 FF
        if (uint8Array[0] === 0xFF && uint8Array[1] === 0xD8) {
            return 'jpg';
        }

        // Check for PNG file signature: 89 50 4E 47 0D 0A 1A 0A
        if (uint8Array[0] === 0x89 && uint8Array[1] === 0x50 && uint8Array[2] === 0x4E && uint8Array[3] === 0x47) {
            return 'png';
        }

        return 'unknown';
    }
    public static async run(imageBuffer: ArrayBuffer): Promise<ArrayBuffer> {
        const type = this.isType(imageBuffer);
        const decode = type === 'png' ? pngDecode : jpgDecode;

        let imageData = await decode(imageBuffer);
        let resizedWidth = imageData.width;
        let resizedHeight = imageData.height;
        while (resizedWidth > 2000 || resizedHeight > 2000) {
            resizedWidth = Math.floor(resizedWidth / 2);
            resizedHeight = Math.floor(resizedHeight / 2);
        }
        if (resizedWidth !== imageData.width || resizedHeight !== imageData.height) {
            imageData = await resize(imageData, { width: resizedWidth, height: resizedHeight });
        }
        return await jpgEncode(imageData, { quality: 85 });
    }
}