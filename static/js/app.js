$(function () {
    if ($("#numbers-section").length) {
        loadDashboard();
    }
    $("#search-criteria-delete-button").click(handleDeleteSearchCriteria);

    if ($(".tweet-text").length) {
        makeLinks();
    }

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
            $(location).attr("href", "/view/search");
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
        $("#follower-count .data").text(data.user.followers_count).digits();
        $("#friend-count .data").text(data.user.friend_count).digits();
        $("#follower-gained-count .data").text(data.recent_follower_count).digits();
        $("#follower-lost-count .data").text(data.recent_unfollower_count).digits();
        $("#listed-count .data").text(data.user.listed_count).digits();
        $("#post-count .data").text(data.user.post_count).digits();

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
                    text: 'Daily Followers',
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
                    backgroundColor: 'rgba(127, 201, 143,0.1)',
                    borderColor: 'rgba(127, 201, 143,0.5)',
                    borderWidth: 1
                }, {
                    label: 'Unfollowed',
                    data: Object.values(data.unfollowed_event_series),
                    backgroundColor: 'rgba(206, 149, 166,0.1)',
                    borderColor: 'rgba(206, 149, 166,0.5)',
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                title: {
                    display: true,
                    text: 'Daily Follower Events',
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
                    if (item.length) {
                        var model = item[0]._model;
                        console.log("Date: ", model.label);
                        redirectToDate(model.label);
                    }
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

function makeLinks() {
    var tweetText = $(".tweet-text");
    if (tweetText.length) {
        tweetText.each(
            function () {
                var $words = $(this).text().split(' ');
                for (i in $words) {
                    if ($words[i].indexOf('https://t.co/') == 0) {
                        $words[i] = '<a href="' + $words[i] + '" target="_blank">' + $words[i] + '</a>';
                    }
                }
                $(this).html($words.join(' '));
            }
        );
    }
}

$.fn.digits = function () {
    return this.each(function () {
        $(this).text($(this).text().replace(/(\d)(?=(\d\d\d)+(?!\d))/g, "$1,"));
    })
}