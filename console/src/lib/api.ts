declare global {
    interface Window {
        API_PREFIX?: string;
    }
}

export const getApiPath = (path: string) => {
    const prefix = window.API_PREFIX || '/api';
    // 移除 path 开头的 / 防止双重斜杠
    const cleanPath = path.startsWith('/') ? path.slice(1) : path;
    // 移除 prefix 结尾的 /
    const cleanPrefix = prefix.endsWith('/') ? prefix.slice(0, -1) : prefix;
    return `${cleanPrefix}/${cleanPath}`;
};
