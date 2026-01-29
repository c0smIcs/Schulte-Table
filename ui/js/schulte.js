const evtSource = new EventSource('/timer')

evtSource.onmessage = function (event) {
	const timerElement = document.getElementById('timer')
	if (timerElement) {
		timerElement.textContent = event.data
	}
}

evtSource.onerror = function () {
	evtSource.close()
}

async function sendClick(event, element) {
	event.preventDefault() 

	if (element.classList.contains('done')) return

	const val = element.getAttribute('data-val')

	try {
		const response = await fetch('/click?val=' + val)
		if (!response.ok) return

		const data = await response.json()

		if (data.is_correct === true) {
			element.classList.add('done')

			const nextNumElement = document.getElementById('next-num')
			if (nextNumElement) {
				nextNumElement.textContent = data.next_number
			}

			if (data.status === 'Won!') {
				const infoContainer = document.querySelector('.info')

				infoContainer.innerHTML = `
                    <div class="victory-panel">
                        <span class="final-time">Время: ${data.time_taken}</span>
                        <a href="/restart" class="btn-restart">Играть снова ↻</a>
                    </div>
                `
			}
		}
	} catch (error) {
		console.error(error)
	}
}
