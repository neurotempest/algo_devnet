$(function() {
  $('select[name=my-select]').on('change', onChange);

  function onChange(e) {
    e.preventDefault();
    change();
  }
  change()

  function change() {
    let val = $('select[name=my-select]').val();

    console.log(val);

    if (val === 'item1') {
      $('#item1_div').removeClass('d-none');
      $('#item2_div').addClass('d-none');
    } else if (val === 'item2') {
      $('#item1_div').addClass('d-none');
      $('#item2_div').removeClass('d-none');
    } else {
      $('#item1_div').addClass('d-none');
      $('#item2_div').addClass('d-none');
    }
  }
});
