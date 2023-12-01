
function updateHTML() {

    const main = document.getElementById('maincontent')

    let html = '';

    for (const srv of servers) {

        html += `<a class="server" href="${srv['protocol'].toLowerCase()}://localhost:${srv['port']}">${encodeURIComponent(srv['process'])} @ ${encodeURIComponent(srv['port'])}</a>`

    }
    main.innerHTML = html;

}

document.addEventListener("DOMContentLoaded", updateHTML);