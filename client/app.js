document.getElementById('login-form').addEventListener('submit', function (e) {
    e.preventDefault();

    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    alert('Login successful!');
    window.location.href = 'home.html';
});

document.getElementById('register-link').addEventListener('click', function () {
    alert('Register functionality is not implemented yet!');
});
