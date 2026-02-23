const root = document.documentElement;
let lastFrom;
let lastTo;

function setTheme(theme) {
    root.setAttribute("data-theme", theme);
    localStorage.setItem("theme", theme);
}

document.getElementById("themeToggle").onclick = () => {
    const current = root.getAttribute("data-theme") === "dark" ? "light" : "dark";
    setTheme(current);
    loadHistory(lastFrom, lastTo); // rerender charts
};

function clampUpward(data, threshold) {
    if (!data.length) return data;

    const result = [];
    let prev = data[0].y;

    result.push({ ...data[0] });

    for (let i = 1; i < data.length; i++) {
        const current = data[i].y;

        if (current > prev + threshold) {
            // ignorisi skok nagore
            result.push({ x: data[i].x, y: prev });
        } else {
            result.push({ ...data[i] });
            prev = current;
        }
    }

    return result;
}

function soilToPercent(adc) {
    const min = 1320;   // suvo
    const max = 1880;  // vlažno

    let pct = 100 - (adc - min) / (max - min) * 100;
    pct = Math.min(Math.max(pct, 0), 100); // clamp 0–100
    return pct;
}


// init
const saved = localStorage.getItem("theme") || "dark";
setTheme(saved);
async function loadRecent(limit = 10) {
    const res = await fetch(`/api/recent?limit=${limit}`);
    const rows = await res.json();

    const tbody = document.getElementById("tableBody");
    tbody.innerHTML = "";

    rows.forEach(p => {
        let soilperc = soilToPercent(p.soil).toFixed(1);
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

function cssVar(name) {
    return getComputedStyle(document.documentElement)
        .getPropertyValue(name).trim();
}

function computeRange(data, paddingRatio = 0.1) {
    if (!data.length) return { min: 0, max: 1 };

    let min = Infinity;
    let max = -Infinity;

    for (const p of data) {
        if (p.y < min) min = p.y;
        if (p.y > max) max = p.y;
    }

    if (min === max) {
        return { min: min - 1, max: max + 1 };
    }

    const span = max - min;
    const paddedMin = min - span * paddingRatio;
    const paddedMax = max + span * paddingRatio;

    const magnitude = Math.pow(10, Math.floor(Math.log10(span)));
    const niceStep = magnitude / 2;

    const niceMin = Math.floor(paddedMin / niceStep) * niceStep;
    const niceMax = Math.ceil(paddedMax / niceStep) * niceStep;

    return {
        min: niceMin,
        max: niceMax
    };
}

function makeChart({
    canvasId,
    datasets,
    yLabel,
    yMin,
    yMax,
    autoRange = true,
    instanceKey
}) {
    const allPoints = datasets.flatMap(ds => ds.values);

    const range = autoRange
        ? computeRange(allPoints)
        : { min: yMin, max: yMax };
    const dayNightPlugin = {
        id: 'dayNight',
        beforeDraw(chart) {
            const { ctx, chartArea, scales } = chart;
            if (!chartArea) return;

            const xScale = scales.x;
            const { top, bottom } = chartArea;

            ctx.save();

            const start = xScale.min;
            const end = xScale.max;

            const hourMs = 60 * 60 * 1000;
            let t = start - (start % hourMs);

            while (t < end) {
                const date = new Date(t);
                const hour = date.getHours();

                // noć: 20–06
                const isNight = hour >= 20 || hour < 6;

                if (isNight) {
                    const x1 = xScale.getPixelForValue(t);
                    const x2 = xScale.getPixelForValue(t + hourMs);

                    ctx.fillStyle = 'rgba(0, 0, 0, 0.2)';
                    ctx.fillRect(x1, top, x2 - x1, bottom - top);
                }

                t += hourMs;
            }
            ctx.restore();
        }
    };
    const ctx = document.getElementById(canvasId).getContext("2d");

    if (window[instanceKey]) {
        window[instanceKey].destroy();
    }

    window[instanceKey] = new Chart(ctx, {
        type: "line",
        data: {
            datasets: datasets.map(ds => ({
                label: ds.label,
                data: ds.values,
                borderColor: cssVar(ds.color),
                borderWidth: 2,
                tension: 0.2,
                pointRadius: 0,
                pointHoverRadius: 0
            }))
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                x: {
                    type: "time",
                    time: {
                        unit: "hour"
                    },
                    ticks: {
                        maxTicksLimit: 6,
                        color: cssVar("--text")
                    },
                    grid: {
                        display: false
                    }
                },
                y: {
                    title: {
                        display: true,
                        text: yLabel
                    },
                    min: range.min,
                    max: range.max,
                    grid: {
                        color: cssVar("--grid")
                    },
                    ticks: {
                        color: cssVar("--text")

                    }
                }
            }
        },
        plugins: [dayNightPlugin]
    });
}

flatpickr.localize(flatpickr.l10ns.sr);

const fromPicker = flatpickr("#fromDate", {
    dateFormat: "d.m.Y.",
    defaultDate: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000)
});

const toPicker = flatpickr("#toDate", {
    dateFormat: "d.m.Y.",
    defaultDate: new Date()
});

async function loadHistory(from, to) {
    const res = await fetch(
        `/api/range/combined?from=${from}&to=${to}`
    );
    const payload = await res.json();
    const rows = payload.data;

    if (!rows || rows.length === 0) return;

    const tempData = rows.map(p => ({ x: p.ts, y: p.temp_avg }));
    const humData = rows.map(p => ({ x: p.ts, y: p.hum_avg }));
    const lightData = rows.map(p => ({ x: p.ts, y: p.light_max }));
    let soilData = rows.map(p => ({ x: p.ts, y: soilToPercent(p.soil_avg) }));
    soilData = clampUpward(soilData, 8);

    console.log(soilData);

    luxon.Settings.defaultLocale = "sr-Latn-RS";

    makeChart({
        canvasId: "tempChart",
        yLabel: "temp [°C]",
        yMin: 0,
        yMax: 40,
        instanceKey: "tempChartInstance",
        datasets: [
            {
                label: "Temperature",
                values: tempData,
                color: "--temp"
            }
        ]
    });

    makeChart({
        canvasId: "humChart",
        label: "Humidity",
        yLabel: "rHum [%]",
        yMin: 0,
        yMax: 100,
        instanceKey: "humChartInstance",
        autoRange: false,
        datasets: [
            {
                label: "Air humidity",
                values: humData,
                color: "--hum"
            },
            {
                label: "Soil moisture",
                values: soilData,
                color: "--soil"
            }
        ]
    });

    makeChart({
        canvasId: "lightChart",
        yLabel: "Light",
        yMin: 0,
        yMax: 4095,
        instanceKey: "lightChartInstance",
        autoRange: false,
        datasets: [
            {
                label: "Light",
                values: lightData,
                color: "--light"
            }
        ]
    });
}

async function loadLatest() {
    const res = await fetch("/api/latest");
    const data = await res.json();

    // --- GAUGES ---
    const temp = data.t;
    document.getElementById("tempValue").textContent = temp.toFixed(1);
    document.getElementById("tempBar").style.width =
        Math.min(Math.max(temp / 40 * 100, 0), 100) + "%";

    const hum = data.rh;
    document.getElementById("humValue").textContent = hum.toFixed(1);
    document.getElementById("humBar").style.width =
        Math.min(Math.max(hum, 0), 100) + "%";

    const soil = soilToPercent(data.soil);
    document.getElementById("soilValue").textContent = soil.toFixed(0);
    document.getElementById("soilBar").style.width =
        Math.min(Math.max(soil, 0), 100) + "%";

    const light = data.light;
    document.getElementById("lightValue").textContent = light.toFixed(0);
    document.getElementById("lightBar").style.width =
        Math.min(Math.max(light / 4095 * 100, 0), 100) + "%";
}

// initial load
loadLatest();
loadRecent();

// default history: last 24h
const now = Date.now();
lastFrom = now - 24 * 60 * 60 * 1000;
lastTo = now;

loadHistory(now - 24 * 60 * 60 * 1000, now);

setInterval(() => {
    loadLatest();
    loadRecent();
}, 2 * 60 * 1000); // 2 min

document.getElementById("applyBtn").addEventListener("click", () => {
    const range = dayRangeToTs();

    if (!range || range.from >= range.to) {
        alert("Invalid date range");
        return;
    }

    loadHistory(range.from, range.to);
});

function dayRangeToTs() {
    const from = fromPicker.selectedDates[0];
    const to = toPicker.selectedDates[0];

    if (!from || !to) return null;

    const fromTs = new Date(from);
    fromTs.setHours(0, 0, 0, 0);

    const toTs = new Date(to);
    toTs.setHours(23, 59, 59, 999);

    return {
        from: fromTs.getTime(),
        to: toTs.getTime()
    };
}

document.addEventListener("DOMContentLoaded", () => {
    if (document.getElementById("tableBody")) {
        loadRecent();
    }
});
