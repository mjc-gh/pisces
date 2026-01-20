// Get references to the elements
const passwordToggle = document.getElementById('password-toggle');
const passwordInput = document.getElementById('password');

// Add click event listener
passwordToggle.addEventListener('click', function(event) {
  // Prevent any default behavior
  event.preventDefault();

  // Toggle the input type
  if (passwordInput.type === 'password') {
    passwordInput.type = 'text';
  } else {
    passwordInput.type = 'password';
  }
});
