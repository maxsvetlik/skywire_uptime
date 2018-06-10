$("#key-submission").submit(function (e) {
  e.preventDefault();
  console.log("Method")
  publicKey = document.getElementById('public_key_input').value;

  $.ajax({
    url: '/search',
    type: 'POST',
    data: {
      'publicKey': publicKey,
    },
    success: function(result) {
      location.href="/?" + publicKey.replace(/\s+/, "") 
    }
  })
});
