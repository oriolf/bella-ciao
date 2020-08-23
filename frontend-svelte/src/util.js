export function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

export async function submitForm(event, url, paramsFunc) {
    event.preventDefault();
    let data = new FormData(event.target);
    let json = {};
    data.forEach(function (v, k) {
        json[k] = v;
    });
    let validationErrors = [];
    if (paramsFunc !== undefined) {
        ({ json, validationErrors } = paramsFunc(json));
    }
    if (validationErrors.length > 0) {
        return { ok: false, validationErrors: validationErrors };
    }
    return await fetch(url, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify(json),
    });
}

export function validateNonEmpty(name) {
    return function (values) {
        if (values[name] === "") {
            return [{ name: name, error: "is empty but should have a value" }];
        }
        return [];
    }
}

export function validationFuncs(funcs) {
    return function (values) {
        let errors = [];
        for (let f of funcs) {
            let errs = f(values);
            errors = errors.concat(errs);
        }
        return errors;
    }
}

export function paramsFunc(valFunc, jsonFunc) {
    return function (json) {
        let errors = [];
        if (valFunc !== undefined) {
            errors = valFunc(json);
        }
        if (jsonFunc !== undefined) {
            json = jsonFunc(json);
        }
        return { json: json, validationErrors: errors };
    }
}

export function formatValidationError(err) {
    return `Field "${err.name}" ${err.error}`;
}