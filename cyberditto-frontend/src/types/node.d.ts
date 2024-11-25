declare namespace NodeJS {
    interface Timeout {
        _idleTimeout: number;
        _idlePrev: unknown;
        _idleNext: unknown;
        _idleStart: number;
        _onTimeout: () => void;
        _timerArgs: unknown[];
        _repeat: number | null;
    }
}