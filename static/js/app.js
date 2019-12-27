$(function () {
    if ($("#numbers-section").length) {
        loadDashboard();
    }
    $("#search-criteria-delete-button").click(handleDeleteSearchCriteria);
});

function handleDeleteSearchCriteria(e) {
    e.preventDefault();
    var cid = $(this).attr("data-id");
    console.log("deleting: " + cid);
    return $.ajax({
        url: "/data/search/" + cid,
        type: 'DELETE',
        success: function (result) {
            console.log("Delete success: ", result);
            $(location).attr("href", "/search");
        },
        error: function (err) {
            console.log("Delete err: ", err);
        },
    });
}


function loadDashboard() {

    $(".after-load").hide();

    $.get("/data/view", function (data) {
        console.log(data);

        // numbers
        $("#follower-count .data").text(data.user.followers_count);
        $("#following-count .data").text(data.user.following_count);
        $("#follower-gained-count .data").text(data.recent_follower_count);
        $("#follower-lost-count .data").text(data.recent_unfollower_count);
        $("#favorites-count .data").text(data.user.fave_count);

        $("#meta-name").text(data.user.username);
        $("#meta-updated-on").text(toLongTime(data.updated_on));

        $(".wait-load").hide();
        $(".after-load").show();

        // follower count chart
        var followerChart = new Chart($("#follower-count-series")[0].getContext("2d"), {
            type: 'bar',
            data: {
                labels: Object.keys(data.follower_count_series),
                datasets: [{
                    label: 'Count',
                    data: Object.values(data.follower_count_series),
                    backgroundColor: 'rgba(109, 110, 110, 0.4)',
                    borderColor: 'rgba(109, 110, 110, 1)',
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                title: {
                    display: true,
                    text: 'Daily Follower Count (UTC)',
                    fontColor: 'rgba(250, 250, 250, 0.5)',
                    fontSize: 16,
                },
                legend: {
                    display: false
                },
                scales: {
                    yAxes: [
                        {
                            ticks: {
                                beginAtZero: false,
                                fontColor: 'rgba(250, 250, 250, 0.5)',
                                fontSize: 14
                            }
                        }
                    ],
                    xAxes: [
                        {
                            ticks: {
                                fontColor: 'rgba(250, 250, 250, 0.5)',
                                fontSize: 14
                            }
                        }
                    ]
                },
                onClick: (evt, item) => {
                    var model = item[0]._model;
                    console.log("Date: ", model.label);
                    redirectToDate(model.label);
                }
            }
        });


        // follower count chart
        var followerChart = new Chart($("#follower-event-series")[0].getContext("2d"), {
            type: 'bar',
            data: {
                labels: Object.keys(data.followed_event_series),
                datasets: [{
                    label: 'Followed',
                    data: Object.values(data.followed_event_series),
                    backgroundColor: 'rgba(127, 201, 143,0.2)',
                    borderColor: 'rgba(127, 201, 143,0.6)',
                    borderWidth: 1
                }, {
                    label: 'Unfollowed',
                    data: Object.values(data.unfollowed_event_series),
                    backgroundColor: 'rgba(206, 149, 166,0.2)',
                    borderColor: 'rgba(206, 149, 166,0.6)',
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                title: {
                    display: true,
                    text: 'Daily Follower Events (UTC)',
                    fontColor: 'rgba(250, 250, 250, 0.5)',
                    fontSize: 16,
                },
                legend: {
                    display: false
                },
                scales: {
                    yAxes: [
                        {
                            ticks: {
                                beginAtZero: true,
                                fontColor: 'rgba(250, 250, 250, 0.5)',
                                fontSize: 14
                            }
                        }
                    ],
                    xAxes: [
                        {
                            ticks: {
                                fontColor: 'rgba(250, 250, 250, 0.5)',
                                fontSize: 14
                            }
                        }
                    ]
                },
                onClick: (evt, item) => {
                    var model = item[0]._model;
                    console.log("Date: ", model.label);
                    redirectToDate(model.label);
                }
            }
        });

    });
}

function redirectToDate(d) {
    $(location).attr("href", "/view/day/" + d);
}

function toLongTime(v) {
    var ts = new Date(v)
    return ts.toTimeString()
}
