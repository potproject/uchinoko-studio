const SESSION_STORAGE_PREFIX = "chat-session-order";

const buildStorageKey = (ownerId: string, characterId: string) => `${SESSION_STORAGE_PREFIX}:${ownerId}:${characterId}`;

export const getDefaultChatSessionId = (ownerId: string) => ownerId;

export const createChatSessionId = (ownerId: string) => {
    const randomId = globalThis.crypto?.randomUUID?.() ?? `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
    return `${ownerId}:${randomId}`;
};

export const buildSessionQuery = (ownerId: string, sessionId: string) => {
    if (sessionId === ownerId) {
        return "";
    }
    const params = new URLSearchParams({ sessionId });
    return `?${params.toString()}`;
};

export const getStoredChatSessionOrder = (ownerId: string, characterId: string) => {
    const raw = localStorage.getItem(buildStorageKey(ownerId, characterId));
    if (!raw) {
        return [] as string[];
    }
    try {
        const parsed = JSON.parse(raw);
        if (!Array.isArray(parsed)) {
            return [];
        }
        return parsed.filter((value): value is string => typeof value === "string");
    } catch (e) {
        console.error(e);
        return [];
    }
};

export const setStoredChatSessionOrder = (ownerId: string, characterId: string, sessionIds: string[]) => {
    localStorage.setItem(buildStorageKey(ownerId, characterId), JSON.stringify(sessionIds));
};

export const touchStoredChatSession = (ownerId: string, characterId: string, sessionId: string) => {
    const next = [sessionId, ...getStoredChatSessionOrder(ownerId, characterId).filter((id) => id !== sessionId)];
    setStoredChatSessionOrder(ownerId, characterId, next);
    return next;
};

export const removeStoredChatSession = (ownerId: string, characterId: string, sessionId: string) => {
    const next = getStoredChatSessionOrder(ownerId, characterId).filter((id) => id !== sessionId);
    setStoredChatSessionOrder(ownerId, characterId, next);
    return next;
};
