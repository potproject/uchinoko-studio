class VolumeProcessor extends AudioWorkletProcessor {
    constructor(options) {
        super(options);
        this.threshold = options.processorOptions.threshold;
        this.sampleRate = options.processorOptions.sampleRate;
        this.silentCount = 0;
        this.slientThreshold = 1;
        this.running = false;
        this.port.postMessage({ speak: false });
    }

    process(inputs, outputs, parameters) {
        const input = inputs[0];
        let sum = 0;
        let volume = 0;

        // 最初のチャンネルのみを考慮する
        if (input.length > 0) {
            const samples = input[0];
            for (let i = 0; i < samples.length; i++) {
                sum += samples[i] * samples[i];
            }
            volume = Math.sqrt(sum / samples.length);
            // 閾値を超えたかどうかをメインスレッドに通知
            if (volume > this.threshold) {
                outputs = inputs;
                if(this.running == false){
                    this.port.postMessage({ speak: true });
                    this.running = true;
                }
                this.silentCount = 0;
            }else{
                if(this.running){
                    this.silentCount += samples.length;
                    if(this.silentCount > this.sampleRate * this.slientThreshold){
                        this.port.postMessage({ speak: false });
                        this.running = false;
                        this.silentCount = 0;
                    }
                }
            }
        }

        return true;
    }
}

registerProcessor('volume-processor', VolumeProcessor);