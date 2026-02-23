function soilToPercent(adc) {

    const min = 1320;
    const max = 1880;

    let pct = 100 - (adc - min) / (max - min) * 100;
    pct = Math.min(Math.max(pct, 0), 100);

    return pct;
}

async function loadRecent(limit = 200) {

    const res = await fetch(`/api/recent?limit=${limit}`);
    const rows = await res.json();

    const tbody = document.getElementById("tableBody");
    tbody.innerHTML = "";

    rows.forEach(p => {

        const soilperc = soilToPercent(p.soil).toFixed(1);

        const tr = document.createElement("tr");

        tr.innerHTML = `
            <td>${new Date(p.ts).toLocaleString("sr-Latn-RS")}</td>
            <td>${p.t.toFixed(1)}</td>
            <td>${p.rh.toFixed(1)}</td>
            <td>${soilperc}</td>
            <td>${p.light}</td>
        `;

        tbody.appendChild(tr);
    });
}

document.addEventListener("DOMContentLoaded", () => {
    loadRecent(200);
});
