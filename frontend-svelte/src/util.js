export function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

export function extractFormValues(event) {
    let data = new FormData(event.target);
    let json = {};
    data.forEach(function (v, k) {
        json[k] = v;
    });
    return json;
}

export function validationFuncs(funcs) {
    return function (values) {
        let errors = {};
        for (let f of funcs) {
            let errs = f(values);
            for (var k in errs) {
                errors[k] = errs[k];
            }
        }
        return errors;
    }
}

export async function submitForm(url, values, jsonFunc) {
    if (jsonFunc !== undefined) {
        values = jsonFunc(values);
    }
    return await fetch(url, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify(values),
    });
}