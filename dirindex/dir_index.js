// SPDX-FileCopyrightText: 2026 The SonicWeb contributors.
// SPDX-License-Identifier: MPL-2.0

function sortTable(columnIndex) {
	const table = document.querySelector(".directoryListing");
	const tbody = table.tBodies[0];
	const headerCells = table.querySelectorAll("th");
	const currentHeader = headerCells[columnIndex];

	// get current state
	const currentSort = currentHeader.getAttribute("aria-sort");
	const newSort = currentSort === "ascending" ? "descending" : "ascending";
	const direction = newSort === "ascending" ? 1 : -1;

	// reset all headers
	headerCells.forEach(th => th.setAttribute("aria-sort", "none"));

	// sort lines
	const rows = Array.from(tbody.rows);
	const parentRow = rows.find(r => {
		const link = r.cells[0].querySelector('a');
		return link && link.textContent.trim() === "..";
	});
	const dataRows = rows.filter(r => r !== parentRow);

	dataRows.sort((a, b) => {
		// 1. keep folders up first
		const aIsDir = a.cells[0].querySelector('.icon').textContent === "📁";
		const bIsDir = b.cells[0].querySelector('.icon').textContent === "📁";
		if (aIsDir && !bIsDir) return -1;
		if (!aIsDir && bIsDir) return 1;

		// 2. extract values (prefer datetime)
		const aTime = a.cells[columnIndex].querySelector('time');
		const bTime = b.cells[columnIndex].querySelector('time');

		let aVal, bVal;
		if (aTime && bTime) {
			aVal = aTime.getAttribute('datetime');
			bVal = bTime.getAttribute('datetime');
		} else {
			aVal = a.cells[columnIndex].textContent.trim();
			bVal = b.cells[columnIndex].textContent.trim();
		}

		return aVal.localeCompare(bVal, undefined, {numeric: true, sensitivity: 'base'}) * direction;
	})

	// set new sort order
	currentHeader.setAttribute("aria-sort", newSort);

	// recreate table
	if (parentRow) tbody.appendChild(parentRow);
	dataRows.forEach(row => tbody.appendChild(row));
}

function updateTimeValues() {
	// get all <time> elements of class "auto-date"
	const timeElements = document.querySelectorAll('time.auto-date');

	timeElements.forEach(el => {
		const isoStr = el.getAttribute('datetime');
		if (isoStr) {
			const date = new Date(isoStr);
			// check if date is valid
			if (!isNaN(date.getTime())) {
				el.textContent = date.toLocaleString( undefined, {
					year: 'numeric',
					month: '2-digit',
					day: '2-digit',
					hour: '2-digit',
					minute: '2-digit',
					second: '2-digit'
				});
			}
		}
	});
}

// single call when page is finished loading
window.addEventListener('DOMContentLoaded', () => {
	document.body.classList.add("js-enabled");
	updateTimeValues();
});