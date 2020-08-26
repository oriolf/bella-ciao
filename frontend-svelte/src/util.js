export function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

export function extractFormValuesJSON(event) {
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

export async function get(url, sortFunc) {
    let res = await fetch(url);
    if (!res.ok) {
        throw new Error("Could not get files");
    }
    let r = (await res.json())
    if (sortFunc) {
        r = r.sort(sortFunc);
    }
    return r;
}

export async function submitForm(url, form) {
    return await fetch(url, {
        method: "POST",
        // do not set content type so client detects and sets appropriate content-type and boundary: https://stackoverflow.com/questions/39280438/fetch-missing-boundary-in-multipart-form-data-post
        headers: {},
        body: new FormData(form)
    });
}

export async function submitFormJSON(url, values, jsonFunc) {
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

export function sortByField(field) {
    return (a, b) => {
        if (a[field] < b[field]) {
            return -1;
        } else if (a[field] > b[field]) {
            return 1;
        }
        return 0;
    }
}