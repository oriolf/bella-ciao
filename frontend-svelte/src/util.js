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
    if (paramsFunc !== undefined) {
        json = paramsFunc(json);
    }
    return await fetch(url, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify(json),
    });
}