
$("#key-submission").submit(function (e) {
  e.preventDefault();
  publicKey = document.getElementById('public_key_input').value;

  $.ajax({
    url: '/nodesearch',
    type: 'POST',
    data: {
      'publicKey': publicKey,
    },
    success: function(result) {
      if (result.startsWith("Error")) {
        error_modal("Oh boy...", result);
      } else {
        location.href='/node?msg=searching';
      }
    }
  })
});
