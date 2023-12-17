<script lang="ts">
    /* @ts-ignore */
    import Start from "../component/start.svelte";
    import Chat from "../component/chat.svelte";

    let route: "start" | "chat" = "start";

    const changeRoute = (newRoute: "start" | "chat") => {
        route = newRoute;
    }
</script>

<main class="bg-gradient-to-r from-cyan-100 to-blue-400 w-screen h-screen flex justify-center items-center">
    {#if route === "start"}
    <Start on:start={() => {
        changeRoute("chat");
    }} />
    {/if}
    {#if route === "chat"}
    <Chat />
    {/if}
</main>
