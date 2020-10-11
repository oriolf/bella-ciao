export type User = {
    id: number,
    name: string,
    unique_id: string,
    email: string,
    role: "none" | "validated" | "admin",
    has_voted: boolean,
    messages: UserMessage[],
    files: UserFile[]
};

export type UserMessage = {
    id: number,
    content: string,
    solved: boolean
}

export type UserFile = {
    id: number,
    name: string,
    description: string
}

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

export type FormParams = {
    name: string,
    values?: string,
    url: string,
    generalError?: string,
    jsonFunc?: JsonFunc,
    fields: FormField[]
}

export type FormField = {
    name: string,
    title?: string,
    type?: string,
    hint: string,
    options?: FormOption[],
    required: boolean,
    errString?: string,
    validate?: ValidateGeneratorFunc
}

export type FormOption = {
    id: string,
    name: string
}

export type SortFunc = (x: any, y: any) => number;
export type JsonFunc = (x: JsonValue) => JsonValue;
export type ValidateGeneratorFunc = (x: string) => ValidateFunc;
export type ValidateFunc = (y: JsonValue) => StringMap;

export type StringMap = { [k: string]: string };
export type JsonValue = string | number | boolean | null | JsonValue[] | { [key: string]: JsonValue };
