export type ChatSessionSummary = {
    sessionId: string;
    title: string;
    preview: string;
    messageCount: number;
    isDefault: boolean;
};

export type ChatSessionList = {
    sessions: ChatSessionSummary[];
};
