
const refresh_delay = 2 * 60 * 1000; // 2min

let last_refresh = now();

function now() { return (new Date()).getTime(); }

function enc(v) { return `${v}`.replace(/[\u00A0-\u9999<>&]/g, i => '&#'+i.charCodeAt(0)+';') }

function updateHTML(servers) {

    const main = document.getElementById('maincontent')

    let html = '';

    for (const srv of servers) {

        html += `<a class="server" href="${srv['protocol'].toLowerCase()}://localhost:${srv['port']}" target="_blank">`;
        html += `<span class="txt_icon"><img src="/icons/microchip-sharp-solid.svg" alt="process"></span>`
        html += `<span class="txt_name">${enc(srv['name'])}</span>`
        html += `<span class="txt_port">${enc(srv['port'])}</span>`
        html += `</a>`;

    }
    main.innerHTML = html;

}

function onVisibilityChange() {
    if (!document.hidden && (now() - last_refresh) > refresh_delay) autoReload();
}

async function autoReload() {
    try {
        document.getElementById('loader').classList.remove('hidden');

        const res = await fetch('/api/v1/server');
        if (res.status !== 200) {
            console.error(`status == ${res.status}`);
            return;
        }

        const data = await res.json();

        updateHTML(data['servers'])

    } catch (err) {

        console.error(err);

    } finally {
        document.getElementById('loader').classList.add('hidden');
    }
}

document.addEventListener("DOMContentLoaded", () => updateHTML(initialServers));

document.addEventListener('visibilitychange', onVisibilityChange);

setInterval(autoReload, refresh_delay);