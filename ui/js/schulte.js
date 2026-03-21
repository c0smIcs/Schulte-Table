const evtSource = new EventSource('/timer')

// Слушаем сообщения от сервера (SSE)
evtSource.onmessage = function (event) {
	const timerElement = document.getElementById('timer')
	const infoContainer = document.querySelector('.info')

	// Если сервер прислал "Timeout" — время вышло!
	if (event.data === 'Timeout') {
		evtSource.close() // Останавливаем поток

		if (infoContainer) {
			// Вставляем блок таймаута автоматически
			infoContainer.innerHTML = `
                <div class="timeout-panel animated-fade-in">
                    <span class="timeout-text">⏰ Время вышло!</span>
                    <a href="/restart" class="btn-timeout">Играть снова ↻</a>
                </div>
            `
		}
		document.querySelector('.grid').style.opacity = '0.5'
		document.querySelector('.grid').style.pointerEvents = 'none'
		return
	}

	// Если игра продолжается, просто обновляем цифры таймера
	if (timerElement) {
		timerElement.textContent = event.data
	}
}

evtSource.onerror = function () {
	evtSource.close()
}

// Функция отправки клика
async function sendClick(event, element) {
	event.preventDefault()

	if (element.classList.contains('done') || document.querySelector('.timeout-panel')) {
		return
	}

	const val = element.getAttribute('data-val')

	try {
		const response = await fetch('/click?val=' + val)
		const data = await response.json()

		// Если сервер ответил, что статус уже не Playing — блокируем интерфейс
		if (data.status === 'Timeout') {
			location.reload() // Или просто вызвать отрисовку таймаута
			return
		}

		if (data.is_correct === true) {
			element.classList.add('done')

			const nextNumElement = document.getElementById('next-num')
			if (nextNumElement) {
				nextNumElement.textContent = data.next_number
			}

			// Если победа
			if (data.status === 'Won!') {
				const infoContainer = document.querySelector('.info')
				infoContainer.innerHTML = `
                    <div class="victory-panel animated-fade-in">
                        <span class="final-time">Время: ${data.time_taken}</span>
                        <a href="/restart" class="btn-restart">Играть снова ↻</a>
                    </div>
                `
			}
		}
	} catch (error) {
		console.error('Ошибка клика:', error)
	}
}
