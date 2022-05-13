
function suggest() {
    var guess = $("#guess").val();
    var url = "/suggest/" + guess.split(" ").join(",");

    console.log(url);

    $.ajax({
    url: url,
    type: "post",
    success: handleSuggestions,
    error: reportError,
    });
}

function handleSuggestions(data) {
    $("#suggestions").text(data.join(" "));
}

function reportError(request) {
    $('#suggestions').html("server error");
}
