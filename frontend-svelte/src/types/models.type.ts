export type User = {
    id: number,
    name: string,
    unique_id: string,
    email: string,
    role: "none" | "validated" | "admin",
    has_voted: boolean
};

export type Candidate = {
    id: number,
    name: string,
    presentation: string,
    image: string,
};

export type Election = {
    id: number,
    name: string,
    start: string,
    end: string,
    count_method: string,
    max_candidates: number,
    min_candidates: number,
    candidates: Candidate[]
}

export type SortFunc = (x: any, y: any) => number;
export type JsonFunc = (x: JsonValue) => JsonValue;

export type JsonValue = string | number | boolean | null | JsonValue[] | { [key: string]: JsonValue };
