export class ScreenCapture {
    private capturePromise: Promise<File> | null = null;
    public stream: MediaStream | null = null;

    constructor() {
      if (!this.isSupported()) {
        console.warn("画面キャプチャAPIがサポートされていません。");
      }
    }

    public async startCapture() {
        if (!this.stream) {
            this.stream = await navigator.mediaDevices.getDisplayMedia({ video: true });
        }
    }

    public async stopCapture() {
        if (this.stream) {
            this.stream.getTracks().forEach(track => track.stop());
            this.stream = null;
        }
    }
  
    public isSupported(): boolean {
      return !!(navigator.mediaDevices && navigator.mediaDevices.getDisplayMedia);
    }
  
    public async capture(): Promise<File> {
      if (this.capturePromise) {
        return this.capturePromise;
      }
  
      this.capturePromise = new Promise<File>(async (resolve, reject) => {
        try {
          if (this.stream?.active === false) {
            reject(new Error("キャプチャが開始されていません。"));
            return;
          }
          const track = this.stream?.getVideoTracks()[0];
          /* @ts-ignore */
          const imageCapture = new ImageCapture(track);
          const bitmap = await imageCapture.grabFrame();
          
          const canvas = document.createElement('canvas');
          canvas.width = bitmap.width;
          canvas.height = bitmap.height;
          const context = canvas.getContext('2d');
          if (!context) {
            throw new Error("キャンバスコンテキストの取得に失敗しました。");
          }
          context.drawImage(bitmap, 0, 0);
          
          canvas.toBlob((blob) => {
            if (!blob) {
              reject(new Error("Blobの生成に失敗しました。"));
              return;
            }
            const file = new File([blob], 'screen-capture.png', { type: 'image/png' });
            resolve(file);
          }, 'image/png');
          
          track?.stop();
        } catch (err) {
          reject(err);
        } finally {
          this.capturePromise = null;
        }
      });
  
      return this.capturePromise;
    }
  }