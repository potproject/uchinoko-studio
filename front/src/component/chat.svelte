<script lang="ts">
    import { onMount } from "svelte";
    import ChatThread from "./chat-thread.svelte";
    import ChatMemory from "./chat-memory.svelte";
    import type { CharacterConfig } from "../types/character";
    import type { GeneralConfig } from "../types/general";
    import type { ChatSessionList, ChatSessionSummary } from "../types/chat";
    import { createChatSessionId, getDefaultChatSessionId, getStoredChatSessionOrder, removeStoredChatSession, setStoredChatSessionOrder, touchStoredChatSession } from "$lib/ChatSession";
    import { getID } from "$lib/GetId";

    export let audio: AudioContext;
    export let media: MediaStream;
    export let selectCharacter: CharacterConfig;
    export let audioOutputDevicesCharacters: string[];
    export let generalConfig: GeneralConfig;

    const ownerId = getID();
    const defaultSessionId = getDefaultChatSessionId(ownerId);

    let currentSessionId = defaultSessionId;
    let sessions: ChatSessionSummary[] = [];
    let sidebarOpen = false;
    let memoryPanelOpen = false;

    const sessionsEndpoint = `/v1/chat/${ownerId}/${selectCharacter.general.id}/sessions`;
    $: memoryEnabled = selectCharacter.memory.enabled;

    const createDraftSession = (sessionId: string): ChatSessionSummary => ({
        sessionId,
        title: "新しいチャット",
        preview: "",
        messageCount: 0,
        isDefault: sessionId === defaultSessionId,
    });

    const reorderSessions = (items: ChatSessionSummary[]) => {
        const storedOrder = getStoredChatSessionOrder(ownerId, selectCharacter.general.id);
        const fallbackOrder = items.map((session) => session.sessionId);
        const mergedOrder = [...storedOrder.filter((id) => fallbackOrder.includes(id)), ...fallbackOrder.filter((id) => !storedOrder.includes(id))];
        setStoredChatSessionOrder(ownerId, selectCharacter.general.id, mergedOrder);

        return [...items].sort((left, right) => {
            const leftIndex = mergedOrder.indexOf(left.sessionId);
            const rightIndex = mergedOrder.indexOf(right.sessionId);
            return leftIndex - rightIndex;
        });
    };

    const ensureCurrentSession = (items: ChatSessionSummary[]) => {
        if (items.some((session) => session.sessionId === currentSessionId)) {
            return items;
        }
        return [createDraftSession(currentSessionId), ...items];
    };

    const loadSessions = async () => {
        const response = await fetch(sessionsEndpoint);
        const data = await response.json() as ChatSessionList;
        sessions = reorderSessions(ensureCurrentSession(data.sessions));
    };

    const selectSession = (sessionId: string) => {
        currentSessionId = sessionId;
        sessions = reorderSessions(ensureCurrentSession(sessions));
        touchStoredChatSession(ownerId, selectCharacter.general.id, sessionId);
        sidebarOpen = false;
    };

    const createSession = () => {
        const sessionId = createChatSessionId(ownerId);
        currentSessionId = sessionId;
        touchStoredChatSession(ownerId, selectCharacter.general.id, sessionId);
        sessions = reorderSessions(ensureCurrentSession(sessions));
        sidebarOpen = false;
    };

    const deleteSession = async (sessionId: string) => {
        if (!window.confirm("このチャットを削除しますか？")) {
            return;
        }

        const query = sessionId === ownerId ? "" : `?${new URLSearchParams({ sessionId }).toString()}`;
        await fetch(`/v1/chat/${ownerId}/${selectCharacter.general.id}${query}`, {
            method: "DELETE",
        });

        removeStoredChatSession(ownerId, selectCharacter.general.id, sessionId);
        const remaining = sessions.filter((session) => session.sessionId !== sessionId);

        if (currentSessionId === sessionId) {
            currentSessionId = remaining[0]?.sessionId ?? defaultSessionId;
        }

        sessions = reorderSessions(ensureCurrentSession(remaining));
    };

    const updateSessionMeta = (summary: ChatSessionSummary) => {
        const next = sessions.filter((session) => session.sessionId !== summary.sessionId);
        next.push(summary);
        touchStoredChatSession(ownerId, selectCharacter.general.id, summary.sessionId);
        sessions = reorderSessions(ensureCurrentSession(next));
    };

    onMount(async () => {
        touchStoredChatSession(ownerId, selectCharacter.general.id, currentSessionId);
        await loadSessions();
    });

    $: if (!memoryEnabled && memoryPanelOpen) {
        memoryPanelOpen = false;
    }
</script>

<div class="w-full h-full relative overflow-hidden">
    {#if sidebarOpen}
        <button class="md:hidden absolute inset-0 bg-slate-950/30 z-20" aria-label="close sidebar" on:click={() => (sidebarOpen = false)}></button>
    {/if}
    {#if memoryEnabled && memoryPanelOpen}
        <button class="absolute inset-0 bg-slate-950/30 z-20" aria-label="close memory panel" on:click={() => (memoryPanelOpen = false)}></button>
    {/if}

    <div class="flex h-full w-full">
        <aside class="absolute left-0 top-0 z-30 flex h-full w-80 max-w-[85vw] flex-col border-r border-slate-200 bg-white/90 backdrop-blur transition-transform duration-200 md:static md:max-w-none md:translate-x-0 {sidebarOpen ? 'translate-x-0' : '-translate-x-full'}">
            <div class="border-b border-slate-200 px-4 py-4">
                <div class="flex items-start justify-between gap-3">
                    <div>
                        <p class="text-xs font-semibold uppercase tracking-[0.24em] text-slate-500">Chats</p>
                        <h2 class="mt-1 text-lg font-bold text-slate-900">{selectCharacter.general.name}</h2>
                    </div>
                    <button class="rounded-full border border-slate-300 px-3 py-2 text-sm font-semibold text-slate-700 transition hover:border-slate-400 hover:bg-slate-100" on:click={createSession}>
                        <i class="las la-plus mr-1"></i>新規
                    </button>
                </div>
            </div>

            <div class="hidden-scrollbar flex-1 overflow-y-auto p-3">
                {#each sessions as session}
                    <div class="mb-2 flex items-start gap-2 rounded-2xl border px-3 py-3 transition {session.sessionId === currentSessionId ? 'border-cyan-400 bg-cyan-50 shadow-sm' : 'border-slate-200 bg-white hover:border-slate-300 hover:bg-slate-50'}">
                        <button class="flex min-w-0 flex-1 items-start gap-3 text-left" on:click={() => selectSession(session.sessionId)}>
                            <div class="mt-1 flex h-9 w-9 shrink-0 items-center justify-center rounded-xl {session.sessionId === currentSessionId ? 'bg-cyan-500 text-white' : 'bg-slate-200 text-slate-600'}">
                                <i class="las la-comments text-lg"></i>
                            </div>
                            <div class="min-w-0 flex-1">
                                <div class="truncate text-sm font-semibold text-slate-900">{session.title}</div>
                                <div class="mt-1 truncate text-xs text-slate-500">{session.preview || "まだメッセージはありません"}</div>
                            </div>
                        </button>
                        <button
                            class="mt-1 rounded-lg p-1 text-slate-400 transition hover:bg-slate-200 hover:text-slate-700"
                            aria-label="delete session"
                            on:click={() => deleteSession(session.sessionId)}
                        >
                            <i class="las la-trash"></i>
                        </button>
                    </div>
                {/each}
            </div>
        </aside>

        <div class="relative flex min-w-0 flex-1 flex-col">
            <div class="absolute left-3 right-3 top-3 z-20 flex items-center justify-between gap-3 md:right-auto">
                <button class="rounded-full bg-white/90 p-3 text-slate-700 shadow-md backdrop-blur md:hidden" on:click={() => (sidebarOpen = true)}>
                    <i class="las la-bars text-xl"></i>
                </button>
                {#if memoryEnabled}
                    <div class="ml-auto flex items-center gap-2">
                        <button class="rounded-full bg-white/90 px-4 py-3 text-sm font-semibold text-slate-700 shadow-md backdrop-blur transition hover:bg-white" on:click={() => (memoryPanelOpen = !memoryPanelOpen)}>
                            <i class={"las mr-2 " + (memoryPanelOpen ? "la-times" : "la-brain")}></i>
                            {memoryPanelOpen ? "閉じる" : "Memory"}
                        </button>
                    </div>
                {/if}
            </div>

            <div class="h-full w-full">
                {#key currentSessionId}
                    <ChatThread
                        audio={audio}
                        media={media}
                        ownerId={ownerId}
                        sessionId={currentSessionId}
                        selectCharacter={selectCharacter}
                        audioOutputDevicesCharacters={audioOutputDevicesCharacters}
                        generalConfig={generalConfig}
                        on:meta={(event) => updateSessionMeta(event.detail)}
                    />
                {/key}
            </div>
        </div>

        {#if memoryEnabled}
            <aside class="absolute right-0 top-0 z-30 flex h-full w-[22rem] max-w-[88vw] flex-col border-l border-slate-200 bg-white/90 shadow-2xl backdrop-blur transition-transform duration-200 {memoryPanelOpen ? 'translate-x-0' : 'translate-x-full'}">
                <ChatMemory
                    ownerId={ownerId}
                    characterId={selectCharacter.general.id}
                    sessionId={currentSessionId}
                    memoryEnabled={memoryEnabled}
                />
            </aside>
        {/if}
    </div>
</div>
