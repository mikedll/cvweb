
document.addEventListener("DOMContentLoaded", () => {

  console.log(`main.js executing`);
  
  const pastResults = document.querySelector('.past-results ul');
  if(pastResults !== null) {
    if(window.results === '') {
      pastResults.appendChild(document.createTextNote("No past results."));
      return;
    }
    
    const results = JSON.parse(window.results);

    if(results.length === 0) {
      const p = document.createElement('p');
      p.appendChild(document.createTextNode('0 results found.'));
      pastResults.closest('div').appendChild(p);
      return;
    }
    
    // console.log(`found ${results.length} results`);
    results.forEach((result) => {
      console.log("result: ", result.uuid);
      const li = document.createElement('li');
      const a = document.createElement('a');
      a.appendChild(document.createTextNode(result.uuid));
      a.href = `/requests/${result.uuid}`;
      li.appendChild(a);
      pastResults.appendChild(li);
    });    
  } else {
    // console.log(`no past results node found`);    
  }

});
