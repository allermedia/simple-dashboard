var Dashboard = React.createClass({displayName: "Dashboard",
  getInitialState: function() {
    return {data: []};
  },
  loadDashboard: function() {
    $.ajax({
      url: this.props.url,
      datatype: 'json',
      success: function(data) {
        this.setState({data: data});
      }.bind(this),
      error: function(xhr, status, err) {
        console.error(this.props.url, status, err.toString());
      }.bind(this)
    });
  },
  componentDidMount: function() {
    this.loadDashboard()
    ws = new WebSocket("ws://"+window.location.host+"/receive");
    console.log("WebSocket opened")
    ws.onmessage = function(evt) {
      var entry = JSON.parse(evt.data)
      var entries = this.state.data
      entries.unshift(entry)

      for (var i = 1; i < entries.length; i++) {
        if (entries[i].name == entry.name) {
          entries.splice(i, 1)
          break;
        }
      }

      this.setState({data: entries})
    }.bind(this)
  },
  render: function() {
    return (
      React.createElement("div", null, 
        React.createElement("h1", null, "Build dashboard"), 
        React.createElement("div", {id: "entry-list"}, 
        React.createElement(BuildEntryList, {data: this.state.data})
        )
      )
    )
  }
});

var BuildEntryList = React.createClass({displayName: "BuildEntryList",
  render: function() {
    var entries = this.props.data.map(function (entry) {
      return (
        React.createElement(Entry, {status: entry.status, name: entry.name})
      );
    });
    return (
      React.createElement("div", {className: "entries"}, 
      entries
      )
    );
  }
});

var Entry = React.createClass({displayName: "Entry", 
  render: function() {
    return (
      React.createElement("div", {className: this.props.status}, 
      React.createElement("h2", null, this.props.name)
      )
    );
  }
});



React.render(
  React.createElement(Dashboard, {url: "/entry/build"}),
  document.getElementById('build-dashboard')
);