<script lang="ts">
    import { browser } from "$app/environment";
    import { onMount } from "svelte";
    import type { MemoryItem, MemoryItemList, SessionSummary } from "../types/character";

    export let ownerId: string;
    export let characterId: string;
    export let sessionId: string;
    export let memoryEnabled = false;

    const characterKindOptions = ["persona_rule", "profile_fact", "instruction_preference"];
    const relationshipKindOptions = ["relationship_fact", "preference", "promise", "ongoing_topic", "instruction_preference", "profile_fact"];
    const kindLabels: Record<string, string> = {
        persona_rule: "人格ルール",
        profile_fact: "プロフィール情報",
        instruction_preference: "指示の好み",
        relationship_fact: "関係性の事実",
        preference: "好み",
        promise: "約束",
        ongoing_topic: "継続中の話題",
    };

    let loading = false;
    let error = "";
    let memoryItems: MemoryItem[] = [];
    let sessionSummary: SessionSummary | null = null;
    let lastLoadedKey = "";
    let mounted = false;
    let creatingScope: "" | "character" | "relationship" = "";
    let savingItemIds: string[] = [];
    let deletingItemIds: string[] = [];

    const formatUpdatedAt = (value: string) => {
        if (!value) {
            return "";
        }
        const date = new Date(value);
        if (Number.isNaN(date.getTime())) {
            return value;
        }
        return date.toLocaleString("ja-JP", {
            month: "short",
            day: "numeric",
            hour: "2-digit",
            minute: "2-digit",
        });
    };

    const updateBusyIds = (ids: string[], id: string, busy: boolean) => {
        if (busy) {
            return ids.includes(id) ? ids : [...ids, id];
        }
        return ids.filter((value) => value !== id);
    };

    const replaceMemoryItem = (item: MemoryItem) => {
        memoryItems = memoryItems.map((value) => value.id === item.id ? item : value);
    };

    const getKindOptions = (item: MemoryItem) => {
        const baseOptions = item.scope === "character" ? characterKindOptions : relationshipKindOptions;
        return baseOptions.includes(item.kind) ? baseOptions : [item.kind, ...baseOptions];
    };

    const getKindLabel = (kind: string) => kindLabels[kind] ?? kind;

    const loadMemory = async (force = false) => {
        if (!browser || !mounted) {
            return;
        }
        const key = `${ownerId}:${characterId}:${sessionId}`;
        if (!force && (!ownerId || !characterId || !sessionId || key === lastLoadedKey)) {
            return;
        }

        loading = true;
        error = "";
        try {
            const encodedOwnerId = encodeURIComponent(ownerId);
            const encodedCharacterId = encodeURIComponent(characterId);
            const encodedSessionId = encodeURIComponent(sessionId);
            const [itemsRes, summaryRes] = await Promise.all([
                fetch(`/v1/memory/${encodedOwnerId}/${encodedCharacterId}/items`),
                fetch(`/v1/memory/${encodedOwnerId}/${encodedCharacterId}/session/${encodedSessionId}/summary`),
            ]);

            if (!itemsRes.ok) {
                throw new Error(`メモリ一覧の取得に失敗しました (${itemsRes.status})`);
            }
            if (!summaryRes.ok) {
                throw new Error(`セッション要約の取得に失敗しました (${summaryRes.status})`);
            }

            const items = await itemsRes.json() as MemoryItemList;
            memoryItems = items.items;
            sessionSummary = await summaryRes.json() as SessionSummary;
            lastLoadedKey = key;
        } catch (e) {
            error = e instanceof Error ? e.message : "メモリ情報の取得に失敗しました";
        } finally {
            loading = false;
        }
    };

    const createMemoryItem = async (scope: "character" | "relationship") => {
        creatingScope = scope;
        error = "";
        try {
            const encodedOwnerId = encodeURIComponent(ownerId);
            const encodedCharacterId = encodeURIComponent(characterId);
            const response = await fetch(`/v1/memory/${encodedOwnerId}/${encodedCharacterId}/items`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    scope,
                    kind: scope === "character" ? "persona_rule" : "relationship_fact",
                    content: "",
                    keywordsText: "",
                    pinned: scope === "character",
                    confidence: 1,
                    salience: 1,
                }),
            });
            if (!response.ok) {
                throw new Error(`メモリの追加に失敗しました (${response.status})`);
            }
            const created = await response.json() as MemoryItem;
            memoryItems = [created, ...memoryItems];
        } catch (e) {
            error = e instanceof Error ? e.message : "メモリの追加に失敗しました";
        } finally {
            creatingScope = "";
        }
    };

    const saveMemoryItem = async (item: MemoryItem) => {
        savingItemIds = updateBusyIds(savingItemIds, item.id, true);
        error = "";
        try {
            const response = await fetch(`/v1/memory/item/${encodeURIComponent(item.id)}`, {
                method: "PATCH",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(item),
            });
            if (!response.ok) {
                throw new Error(`メモリの保存に失敗しました (${response.status})`);
            }
            const updated = await response.json() as MemoryItem;
            replaceMemoryItem(updated);
        } catch (e) {
            error = e instanceof Error ? e.message : "メモリの保存に失敗しました";
        } finally {
            savingItemIds = updateBusyIds(savingItemIds, item.id, false);
        }
    };

    const deleteMemoryItem = async (item: MemoryItem) => {
        if (!window.confirm("このメモリを削除しますか？")) {
            return;
        }
        deletingItemIds = updateBusyIds(deletingItemIds, item.id, true);
        error = "";
        try {
            const response = await fetch(`/v1/memory/item/${encodeURIComponent(item.id)}`, {
                method: "DELETE",
            });
            if (!response.ok) {
                throw new Error(`メモリの削除に失敗しました (${response.status})`);
            }
            memoryItems = memoryItems.filter((value) => value.id !== item.id);
        } catch (e) {
            error = e instanceof Error ? e.message : "メモリの削除に失敗しました";
        } finally {
            deletingItemIds = updateBusyIds(deletingItemIds, item.id, false);
        }
    };

    onMount(() => {
        mounted = true;
        void loadMemory();
    });

    $: if (mounted && ownerId && characterId && sessionId) {
        void loadMemory();
    }

    $: characterMemories = memoryItems.filter((item) => item.scope === "character");
    $: relationshipMemories = memoryItems.filter((item) => item.scope === "relationship");
</script>

<div class="flex h-full flex-col bg-white/85">
    <div class="border-b border-slate-200 px-4 py-4">
        <div class="flex items-start justify-between gap-3">
            <div>
                <p class="text-xs font-semibold uppercase tracking-[0.24em] text-slate-500">Memory</p>
                <h2 class="mt-1 text-lg font-bold text-slate-900">参照中のメモリ</h2>
                <p class="mt-1 text-xs text-slate-500">現在のキャラクターとセッションに紐づく要約と記憶です。</p>
            </div>
            <button class="px-3 py-2 text-sm font-semibold text-slate-700 transition hover:border-slate-400 hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-60" on:click={() => loadMemory(true)} disabled={loading}>
                <i class={"las mr-1 " + (loading ? "la-spinner animate-spin" : "la-sync")}></i>
            </button>
        </div>
    </div>

    <div class="hidden-scrollbar flex-1 space-y-4 overflow-y-auto p-4">
        {#if !memoryEnabled}
            <div class="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
                このキャラクターでは Memory 機能が無効です。設定画面で有効化すると、ここに保持された内容が表示されます。
            </div>
        {/if}

        {#if error}
            <div class="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">{error}</div>
        {/if}

        <section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
            <div class="mb-2 flex items-center justify-between gap-2">
                <h3 class="text-sm font-semibold text-slate-900">セッション要約</h3>
                {#if sessionSummary?.updatedAt}
                    <span class="text-xs text-slate-500">{formatUpdatedAt(sessionSummary.updatedAt)}</span>
                {/if}
            </div>
            <p class="whitespace-pre-wrap text-sm leading-6 text-slate-700">
                {sessionSummary?.summary?.trim() || "まだ要約はありません。会話が蓄積されるとここに表示されます。"}
            </p>
        </section>

        <section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
            <div class="mb-3 flex items-center justify-between gap-2">
                <div class="flex items-center gap-2">
                    <h3 class="text-sm font-semibold text-slate-900">キャラクターメモリ</h3>
                    <span class="rounded-full bg-slate-100 px-2 py-1 text-xs font-semibold text-slate-600">{characterMemories.length}件</span>
                </div>
                <button class="rounded-full border border-cyan-300 px-3 py-2 text-xs font-semibold text-cyan-700 transition hover:border-cyan-400 hover:bg-cyan-50 disabled:cursor-not-allowed disabled:opacity-60" on:click={() => createMemoryItem("character")} disabled={creatingScope !== ""}>
                    <i class={"las mr-1 " + (creatingScope === "character" ? "la-spinner animate-spin" : "la-plus")}></i>追加
                </button>
            </div>
            {#if characterMemories.length === 0}
                <p class="text-sm text-slate-500">固定のキャラクターメモリはまだありません。</p>
            {:else}
                <div class="space-y-3">
                    {#each characterMemories as item}
                        <article class="rounded-2xl border border-slate-200 bg-slate-50 px-3 py-3">
                            <div class="mb-3 flex flex-wrap items-center gap-2">
                                <select class="rounded-full border border-slate-300 bg-white px-3 py-1 text-xs font-semibold text-slate-700" bind:value={item.kind}>
                                    {#each getKindOptions(item) as option}
                                        <option value={option}>{getKindLabel(option)}</option>
                                    {/each}
                                </select>
                                <label class="inline-flex items-center gap-1 rounded-full bg-amber-100 px-2 py-1 text-xs font-semibold text-amber-700">
                                    <input type="checkbox" bind:checked={item.pinned} />
                                    pinned
                                </label>
                                {#if item.source}
                                    <span class="rounded-full bg-slate-200 px-2 py-1 text-xs font-semibold text-slate-600">{item.source}</span>
                                {/if}
                                <span class="ml-auto text-xs text-slate-500">{formatUpdatedAt(item.updatedAt)}</span>
                            </div>
                            <textarea class="min-h-[6rem] w-full rounded-2xl border border-slate-200 bg-white px-3 py-3 text-sm leading-6 text-slate-800" bind:value={item.content} placeholder="memory content"></textarea>
                            <input class="mt-2 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700" bind:value={item.keywordsText} placeholder="keywords" />
                            <div class="mt-3 grid grid-cols-2 gap-2">
                                <label class="block text-xs text-slate-500">
                                    confidence
                                    <input class="mt-1 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700" type="number" min="0" max="1" step="0.01" bind:value={item.confidence} />
                                </label>
                                <label class="block text-xs text-slate-500">
                                    salience
                                    <input class="mt-1 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700" type="number" min="0" max="1" step="0.01" bind:value={item.salience} />
                                </label>
                            </div>
                            <div class="mt-3 flex items-center justify-end gap-2">
                                <button class="rounded-full border border-slate-300 px-3 py-2 text-sm font-semibold text-slate-700 transition hover:border-slate-400 hover:bg-white disabled:cursor-not-allowed disabled:opacity-60" on:click={() => saveMemoryItem(item)} disabled={savingItemIds.includes(item.id) || deletingItemIds.includes(item.id)}>
                                    <i class={"las mr-1 " + (savingItemIds.includes(item.id) ? "la-spinner animate-spin" : "la-save")}></i>保存
                                </button>
                                <button class="rounded-full border border-rose-300 px-3 py-2 text-sm font-semibold text-rose-700 transition hover:border-rose-400 hover:bg-rose-50 disabled:cursor-not-allowed disabled:opacity-60" on:click={() => deleteMemoryItem(item)} disabled={deletingItemIds.includes(item.id) || savingItemIds.includes(item.id)}>
                                    <i class={"las mr-1 " + (deletingItemIds.includes(item.id) ? "la-spinner animate-spin" : "la-trash")}></i>削除
                                </button>
                            </div>
                        </article>
                    {/each}
                </div>
            {/if}
        </section>

        <section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
            <div class="mb-3 flex items-center justify-between gap-2">
                <div class="flex items-center gap-2">
                    <h3 class="text-sm font-semibold text-slate-900">関係メモリ</h3>
                    <span class="rounded-full bg-slate-100 px-2 py-1 text-xs font-semibold text-slate-600">{relationshipMemories.length}件</span>
                </div>
                <button class="rounded-full border border-emerald-300 px-3 py-2 text-xs font-semibold text-emerald-700 transition hover:border-emerald-400 hover:bg-emerald-50 disabled:cursor-not-allowed disabled:opacity-60" on:click={() => createMemoryItem("relationship")} disabled={creatingScope !== ""}>
                    <i class={"las mr-1 " + (creatingScope === "relationship" ? "la-spinner animate-spin" : "la-plus")}></i>追加
                </button>
            </div>
            {#if relationshipMemories.length === 0}
                <p class="text-sm text-slate-500">ユーザーとの関係メモリはまだありません。</p>
            {:else}
                <div class="space-y-3">
                    {#each relationshipMemories as item}
                        <article class="rounded-2xl border border-slate-200 bg-slate-50 px-3 py-3">
                            <div class="mb-3 flex flex-wrap items-center gap-2">
                                <select class="rounded-full border border-slate-300 bg-white px-3 py-1 text-xs font-semibold text-slate-700" bind:value={item.kind}>
                                    {#each getKindOptions(item) as option}
                                        <option value={option}>{getKindLabel(option)}</option>
                                    {/each}
                                </select>
                                <label class="inline-flex items-center gap-1 rounded-full bg-amber-100 px-2 py-1 text-xs font-semibold text-amber-700">
                                    <input type="checkbox" bind:checked={item.pinned} />
                                    pinned
                                </label>
                                {#if item.source}
                                    <span class="rounded-full bg-slate-200 px-2 py-1 text-xs font-semibold text-slate-600">{item.source}</span>
                                {/if}
                                <span class="ml-auto text-xs text-slate-500">{formatUpdatedAt(item.updatedAt)}</span>
                            </div>
                            <textarea class="min-h-[6rem] w-full rounded-2xl border border-slate-200 bg-white px-3 py-3 text-sm leading-6 text-slate-800" bind:value={item.content} placeholder="memory content"></textarea>
                            <input class="mt-2 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700" bind:value={item.keywordsText} placeholder="keywords" />
                            <div class="mt-3 grid grid-cols-2 gap-2">
                                <label class="block text-xs text-slate-500">
                                    confidence
                                    <input class="mt-1 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700" type="number" min="0" max="1" step="0.01" bind:value={item.confidence} />
                                </label>
                                <label class="block text-xs text-slate-500">
                                    salience
                                    <input class="mt-1 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700" type="number" min="0" max="1" step="0.01" bind:value={item.salience} />
                                </label>
                            </div>
                            <div class="mt-3 flex items-center justify-end gap-2">
                                <button class="rounded-full border border-slate-300 px-3 py-2 text-sm font-semibold text-slate-700 transition hover:border-slate-400 hover:bg-white disabled:cursor-not-allowed disabled:opacity-60" on:click={() => saveMemoryItem(item)} disabled={savingItemIds.includes(item.id) || deletingItemIds.includes(item.id)}>
                                    <i class={"las mr-1 " + (savingItemIds.includes(item.id) ? "la-spinner animate-spin" : "la-save")}></i>保存
                                </button>
                                <button class="rounded-full border border-rose-300 px-3 py-2 text-sm font-semibold text-rose-700 transition hover:border-rose-400 hover:bg-rose-50 disabled:cursor-not-allowed disabled:opacity-60" on:click={() => deleteMemoryItem(item)} disabled={deletingItemIds.includes(item.id) || savingItemIds.includes(item.id)}>
                                    <i class={"las mr-1 " + (deletingItemIds.includes(item.id) ? "la-spinner animate-spin" : "la-trash")}></i>削除
                                </button>
                            </div>
                        </article>
                    {/each}
                </div>
            {/if}
        </section>
    </div>
</div>
