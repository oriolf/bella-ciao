export type User = {
    id: number,
    name: string,
    unique_id: string,
    email: string,
    role: "none" | "validated" | "admin"
};

export type Candidate = {
    id: number,
    name: string,
    presentation: string,
    image: string,
};

export type SortFunc = (x: any, y: any) => number;
export type JsonFunc = (x: JsonValue) => JsonValue;

export type JsonValue = string | number | boolean | null | JsonValue[] | { [key: string]: JsonValue };
