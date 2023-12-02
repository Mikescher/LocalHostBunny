
function enc(v) { return `${v}`.replace(/[\u00A0-\u9999<>&]/g, i => '&#'+i.charCodeAt(0)+';') }

function updateHTML() {

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

document.addEventListener("DOMContentLoaded", updateHTML);