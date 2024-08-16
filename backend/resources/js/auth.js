const swapForms = () => {
    const toggle = document.querySelector('.toggle');
    toggle.classList.toggle('on');
    document.querySelector('.auth').classList.toggle('active');
    document.querySelectorAll('.switch p').forEach(el => el.classList.toggle("underline"));
}

const loginRequest = async (e) => {
    e.preventDefault();
    const form = document.querySelector('#form-sign-in');
    const formData = new FormData(form);
    const data = new URLSearchParams();

    for (const [key, value] of formData) {
        data.append(key, value);
    }

    try {
        const res = await fetch("/auth/login", {
            method: 'POST',
            headers: {
                "Content-Type": "application/x-www-form-urlencoded",
            },
            body: data.toString()
        })

        if (res.status === 200) {
            res.json().then(data => {
                console.log(data);
                if (data.token) {
                    localStorage.setItem('token', data.token);
                    htmx.on("htmx:configRequest", (e)=> {
                        e.detail.headers["Authorization"] = "Bearer " + authToken
                    })
                    window.location.href = '/counter';
                }

            })
        } else {
            const errEl = document.querySelector("#err-sign-in")
            errEl.classList.remove('none')

            errEl.outerHTML = await res.text()
        }
    } catch (err) {
        alert(err);
    }
}

document.querySelector('#sign-in-btn').addEventListener('click', loginRequest);
document.querySelector('.toggle').addEventListener('click', swapForms);