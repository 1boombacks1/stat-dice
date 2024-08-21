htmx.on('htmx:beforeSwap', (e) => {
    if (e.detail.isError) {
        document.querySelector('#error').style.display = 'block';
        e.detail.shouldSwap = true;
        e.detail.target = htmx.find("#error")
    }
});

document.querySelectorAll('.menu-btn').forEach(button => {
    button.addEventListener('click', () => {
        if (!button.querySelector('.menu-icon').classList.contains('deactivated')) {
            // Убираем активный класс со всех кнопок
            document.querySelectorAll('.menu-btn').forEach(btn => {
                const icon = btn.querySelector('.menu-icon');
                if (icon) {
                    icon.classList.remove('leader-active-icon', 'create-lobby-active-icon', 'find-lobby-active-icon', 'completed-active-icon');
                    // Возвращаем иконку в исходное состояние
                    if (icon.id === 'leader-menu-icon') icon.classList.add('leader-icon');
                    if (icon.id === 'create-menu-icon') icon.classList.add('create-lobby-icon');
                    if (icon.id === 'find-menu-icon') icon.classList.add('find-lobby-icon');
                    if (icon.id === 'completed-menu-icon') icon.classList.add('completed-icon');
                }
            });
            // Добавляем активный класс на выбранную кнопку
            const selectedIcon = button.querySelector('.menu-icon');
            if (selectedIcon) {
                if (selectedIcon.id === 'leader-menu-icon') selectedIcon.classList.add('leader-active-icon');
                if (selectedIcon.id === 'create-menu-icon') selectedIcon.classList.add('create-lobby-active-icon');
                if (selectedIcon.id === 'find-menu-icon') selectedIcon.classList.add('find-lobby-active-icon');
                if (selectedIcon.id === 'completed-menu-icon') selectedIcon.classList.add('completed-active-icon');
            }
        }
    });
});

const sidebar = document.querySelector('.side-bar');
const main = document.querySelector('.main');
const toggleButton = document.querySelector('.side-bar-button');

// Обработчик клика по кнопке side-bar
toggleButton.addEventListener('click', function(event) {
    sidebar.classList.toggle('active');
    main.classList.toggle('active');
    event.stopPropagation(); // Предотвращаем всплытие события на документ
});

// Обработчик клика по документу
document.addEventListener('click', function(event) {
    const isClickInsideSidebar = sidebar.contains(event.target);
    const isClickOnToggleButton = toggleButton.contains(event.target);

    // Если клик вне side-bar и не по кнопке, закрываем side-bar
    if (!isClickInsideSidebar && !isClickOnToggleButton) {
        sidebar.classList.remove('active');
        main.classList.remove('active');
    }
});