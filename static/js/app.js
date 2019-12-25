$(function () {
    if ($("#numbers-section").length) {
        runQuery();
    }
    if ($("#search-list-section").length) {
        runSearch();
    }

});

function runSearch() {

    console.log("View: search");
    $(".after-load").hide();
    var listDev = $("#search-list-section").empty();

    $.get("/data/search", function (data) {

        $.each(data, function (i, c) {

            console.log("item[" + i + "]: " + c.id);

            var criteriaLink = $("<a />")
                .data("sc", c)
                .attr("title", "Search Criterion")
                .attr("href", "#")
                .text(c.name)
                .addClass("search-criterion-link")
                .click(function (e) {
                    e.preventDefault();
                    var data = $(this).data("sc");
                    console.log("clicked: " + data.id);
                    loadSearchCriterion(data);
                });

            var execDate = "";
            if (c.since_id > 0) {
                execDate = toLongTime(c.executed_on);
            }
            var criteriaMeta = $("<div />").html("<b>Last executed on:</b> " + execDate);

            $("<div class='search-criterion' />")
                .append(criteriaLink)
                .append(criteriaMeta)
                .appendTo(listDev);
        }); // each

        $("<a />")
            .attr("title", "New Search Criterion")
            .attr("href", "#")
            .text("New Search Criterion")
            .addClass("new-search-criterion-link")
            .click(function (e) {
                e.preventDefault();
                $(".after-load").hide();
                $("#search-new-form")[0].reset();
                $("#search-new-section").show()
            }).appendTo(listDev);

        $("#search-criteria-cancel-button").click(function (e) {
            e.preventDefault();
            $(".after-load").hide();
            $("#search-list-section").show();
        });

        $("#search-criteria-delete-button").click(function (e) {
            e.preventDefault();
            var data = $(this).data("sc");
            if (typeof data === "undefined") {
                runSearch();
                return;
            }
            console.log("deleting: " + data.id);
            return $.ajax({
                url: "/data/search/" + data.id,
                type: 'DELETE',
                success: function (result) {
                    console.log("Delete success: ", result);
                    runSearch();
                },
                error: function (err) {
                    console.log("Delete err: ", err);
                },
            });
        });

        listDev.show();

    });
}

function loadSearchCriterion(data) {

    $(".after-load").hide();

    console.log("View: search detail for " + data.id);
    $("input[name=id]").val(data.id);
    $("input[name=name]").val(data.name);
    $("input[name=value]").val(data.value);

    $("#lang").val(data.lang);

    $("input[name=has_link]").prop("checked", data.has_link);
    $("input[name=include_rt]").prop("checked", data.include_rt);

    $("input[name=post_count_min]").val(data.post_count_min);
    $("input[name=post_count_max]").val(data.post_count_max);

    $("input[name=follower_count_min]").val(data.follower_count_min);
    $("input[name=follower_count_max]").val(data.follower_count_max);

    $("input[name=fave_count_min]").val(data.fave_count_min);
    $("input[name=fave_count_max]").val(data.fave_count_max);

    $("input[name=following_count_min]").val(data.following_count_min);
    $("input[name=following_count_max]").val(data.following_count_max);

    $("input[name=follower_ratio_min]").val(data.follower_ratio_min);
    $("input[name=follower_ratio_max]").val(data.follower_ratio_max);

    $("#search-criteria-delete-button").data("sc", data);

    $("#search-new-section").show()

}

function runQuery() {

    $(".after-load").hide();

    $.get("/data/view", function (data) {
        console.log(data);

        // numbers
        $("#follower-count .data").text(data.user.followers_count);
        $("#following-count .data").text(data.user.following_count);
        $("#follower-gained-count .data").text(data.recent_follower_count);
        $("#follower-lost-count .data").text(data.recent_unfollower_count);
        $("#favorites-count .data").text(data.user.fave_count);

        $(".wait-load").hide();
        $(".after-load").show();

        $("#meta-panel").html("Account: <b>" + data.user.username + "</b>" +
            " | Time period: <b>Last " + data.meta.num_days_period + "days</b>" +
            " | Updated on: <b>" + toLongTime(data.updated_on) + "</b>" +
            " | <a href='/auth/logout'>Log out</a>"
        );

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
                }
            }
        });

        loadUsers($("#follower-list"), data.recent_follower_list);
        loadUsers($("#unfollower-list"), data.recent_unfollower_list);

    });
}

function toShortDate(v) {
    var ts = new Date(v)
    return ts.toISOString().substring(0, 10)
}

function toLongTime(v) {
    var ts = new Date(v)
    return ts.toTimeString()
}

function loadUsers(tbl, list) {
    var lastUser = "";
    $.each(list, function (i, u) {
        if (lastUser == u.username) {
            return true; // continue
        }
        // console.log("User[" + i + "] ID: " + u.id);
        var $info = $("<div class='user-info-detail'>").append(
            $("<div class='user-info-name'>").html("<a href='https://twitter.com/" +
                u.username + "' target='_blank'>" + u.username + "</a> - <b>" + u.name +
                " </b> (<b>Loc:</b> " + u.location +
                " <b>Follow:</b> " + u.followers_count + "/" + u.following_count +
                " <b>Post:</b> " + u.post_count +
                " <b>On:</b> " + toShortDate(u.event_at) + ")"),
            $("<div class='user-info-desc'>").text(u.description),
        );

        var $tr = $("<tr class='user-row'>").append(
            $("<td class='user-img'>").html("<img src='" + u.profile_image + "'/>"),
            $("<td class='user-info'>").append($info)
        ).appendTo(tbl);
        lastUser = u.username;
    });
}