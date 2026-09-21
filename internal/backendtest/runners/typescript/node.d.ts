// The little of the node runtime the runner touches, declared here because @types/node is a package on npm and the
// container has no network. Only the runner needs any of it; the generated scanner and parser touch no runtime API.

declare module "node:fs" {
    /** Reads the whole file. Without an encoding node returns the bytes, which is what the generated scanner takes. */
    export function readFileSync(path: string): Uint8Array;

    /** Writes the whole file, replacing it if it is there. */
    export function writeFileSync(path: string, data: string): void;
}

/** The running process. */
declare const process: {
    /** The command line, the executable and the script ahead of the arguments the runner was given. */
    readonly argv: readonly string[];

    /** Standard error, which everything the runner has to say goes to. */
    readonly stderr: {
        write(message: string): void;
    };
};
