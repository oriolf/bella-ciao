import type { Writable } from "svelte/store";
import type { User, SortFunc, JsonValue, JsonFunc, ValidateFunc, StringMap } from "./types/models.type";

export function sleep(ms: number) {
    return new Promise((resolve) => setTimeout(resolve, ms))
}

export function extractFormValuesJSON(form: HTMLFormElement, multiselectFields: string[]) {
    let data = new FormData(form)
    let json = {};
    data.forEach(function (v, k) {
        if (multiselectFields && multiselectFields.includes(k)) {
            json[k] = data.getAll(k)
        } else {
            json[k] = v
        }
    })
    return json
}

export function validationFuncs(funcs: ValidateFunc[]): ValidateFunc {
    return function (values: JsonValue): StringMap {
        let errors: StringMap = {};
        for (let f of funcs) {
            let errs = f(values);
            for (var k in errs) {
                errors[k] = errs[k]
            }
        }
        return errors;
    }
}

export async function whoami(user: Writable<User>) {
    let res = await fetch('/api/users/whoami')
    if (!res.ok) {
        throw new Error('Not logged in')
    }
    user.set(await res.json() as User)
}

export async function get(url: string, sortFunc: null | SortFunc) {
    let res = await fetch(url)
    if (!res.ok) {
        throw new Error(`Could not get ${url}`)
    }
    let r = await res.json()
    if (sortFunc) {
        r = r.sort(sortFunc)
    }
    return r
}

export async function post(url: string, values: JsonValue) {
    return await fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(values),
    })
}

export async function submitForm(url: string, form: HTMLFormElement) {
    return await fetch(url, {
        method: 'POST',
        // do not set content type so client detects and sets appropriate content-type and boundary: https://stackoverflow.com/questions/39280438/fetch-missing-boundary-in-multipart-form-data-post
        headers: {},
        body: new FormData(form),
    })
}

export async function submitFormJSON(url: string, values: JsonValue, jsonFunc: undefined | JsonFunc) {
    if (jsonFunc !== undefined) {
        values = jsonFunc(values)
    }
    return await fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(values),
    })
}

export function sortByField(field: string): SortFunc {
    return (a, b) => {
        if (a[field] < b[field]) {
            return -1
        } else if (a[field] > b[field]) {
            return 1
        }
        return 0
    }
}

export function formatDate(s: string): string {
    return new Date(s).toLocaleString()
}

export function validateArrayLengthPositive(name: string) {
    return function (values: JsonValue): StringMap {
        if (!values[name] || values[name].length === 0) {
            return {
                [name]: "error"
            };
        }
        return {};
    }
}