<script lang="ts">
    import { createEventDispatcher } from 'svelte';

    const dispatch = createEventDispatcher();
    let micOk: boolean|undefined = undefined;
    let audioOk = false;
    let wsOk = false;
    let start = false;

    let selected = 'bertvits2';

    let slientAudio: HTMLAudioElement;

    let onClick = () => {
        start = true;
        // 音声アンロック
        const audio = new AudioContext();
        const source = audio.createMediaElementSource(slientAudio);
        source.connect(audio.destination);
        slientAudio.play();
        
        slientAudio.onended = () => {
            console.log('start', selected);
            dispatch('start', {
                audio,
                selected
            });
        }
    }

    const checkMic = async () => {
        try {
            await globalThis.navigator.mediaDevices.getUserMedia({ audio: true });
            micOk = true;
        } catch (e) {
            micOk = false;
        }
    }

    const checkAudio = () => {
        // @ts-ignore
        audioOk = !!(globalThis.AudioContext || globalThis.webkitAudioContext);
    }

    const checkWs = () => {
        // @ts-ignore
        wsOk = !!(globalThis.WebSocket || globalThis.MozWebSocket);
    }

    checkMic();
    checkAudio();
    checkWs();
</script>


<main class="bg-gradient-to-r from-cyan-100 to-blue-400 w-screen h-screen flex justify-center items-center">
    <audio src="/audio/silent.mp3" preload="auto" class="hidden" bind:this={slientAudio}></audio>
    <div class="card bg-white shadow-lg rounded-3xl h-auto mx-auto border border-cyan-600 border-opacity-50 border-2 w-96 md:w-1/2 lg:w-1/3 {start ? 'animate-scale-out-horizontal' : 'animate-scale-in-hor-center'}">
        <div class="card-header p-4 flex m-2">
            <h1 class="text-3xl font-bold flex-1">Uchinoko Studio(β)</h1>
        </div>
        <!-- 使用する音声生成サービス -->
        <div class="card-body p-2 m-2">
            <div class="text-gray-500">使用する音声生成サービス/ソフトウェア</div>
            <select bind:value={selected} 
                class="w-full border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring focus:ring-blue-500 focus:border-blue-500"
            >
                <option value="bertvits2">Bert-VITS2(FastAPI)</option>
                <option value="voicevox">VOICEVOX</option>
                <option value="elevenlabs">ELEVENLABS</option>
            </select>
        </div>
        <!-- 概要欄 -->
        <div class="card-body p-3 m-2">
            <h2 class="text-2xl font-bold text-blue-500">
                <i class="las la-info"></i>
                <span>概要</span>
            </h2>
            <p class="text-gray-500">Uchinoko Studioは、AIと音声で通話できることを目指した、Webアプリです。</p>
        </div>
        <!-- 利用規約欄 Textarea -->
        <div class="card-body p-3 m-2">
            <h2 class="text-2xl font-bold text-blue-500">
                <i class="las la-file-alt"></i>
                <span>利用規約</span>
            </h2>
            <textarea class="w-full h-32 p-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring focus:ring-blue-500 focus:border-blue-500" placeholder="利用規約" readonly>
このWebアプリケーションを使用して、起こったいかなる問題についても、開発者は責任を負いません。
音声生成処理に外部のサービスまたはアプリケーションを使用している場合、そのサービスまたはアプリケーションの利用規約に従ってください。
このWebアプリケーションは、Google Chromeを推奨しています。その他のブラウザでは、正常に動作しない可能性があります。
このアプリケーションはオープンソースであり、ソースコードはGitHubで公開されています。ソースコードの利用は、Githubに記載されているライセンスに従ってください。
https://github.com/potproject/uchinoko-studio</textarea>
        </div>
        <div class="card-body p-3 m-2">
            <h2 class="text-2xl font-bold {micOk && audioOk && wsOk ? 'text-green-600' : 'text-red-600'}">
                <i class="las {micOk && audioOk && wsOk ? 'la-check' : 'la-times'}"></i>
                <span>動作チェック</span>
            </h2>
            <div class="flex items-center ml-2 {micOk ? 'text-green-500' : 'text-red-500'}">
                <i class="las la-microphone text-2xl mr-1"></i>
                <span class="">マイク</span>
            </div>
            <div class="flex items-center ml-2 {audioOk ? 'text-green-500' : 'text-red-500'}">
                <i class="las la-volume-up text-2xl mr-1"></i>
                <span class="">スピーカー</span>
            </div>
            <div class="flex items-center ml-2 {wsOk ? 'text-green-500' : 'text-red-500'}">
                <i class="las la-wifi text-2xl mr-1"></i>
                <span class="">ネットワーク</span>
            </div>
            {#if !micOk}
            <div class="text-red-500 text-sm">
                マイクの権限を「許可する」に設定しないと、音声認識ができません。このサービスを利用するには、マイクの権限を許可してください。
            </div>
            {/if}
            {#if !audioOk || !wsOk}
            <div class="text-red-500 text-sm">
                このブラウザでは、このサービスが正常に動作しない可能性があります。推奨するブラウザは、Google ChromeまたはiOS Safariとなります。
            </div>
            {/if}
        <!-- ボタン -->
        <div class="card-footer p-4 flex m-2 justify-center items-center">
            <button on:click={onClick} disabled={!micOk || !audioOk || !wsOk}
            class="btn bg-blue-500 hover:bg-blue-600 text-white font-bold py-2 px-4 rounded-full {micOk && audioOk && wsOk ? 'opacity-100' : 'opacity-50 cursor-not-allowed'}">同意してはじめる</button>
        </div>
    </div>
</main>
