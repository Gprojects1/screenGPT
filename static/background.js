
function createSelectionRectangle(tabId) {
  
  chrome.scripting.executeScript({
      target: { tabId: tabId, allFrames: true },
      function: () => {
          
          console.log("createSelectionRectangle code running"); 
          let startX, startY, rect;
          const body = document.body; 

          
          function createRect() {
              rect = document.createElement('div');
              rect.style.position = 'fixed';
              rect.style.border = '2px dashed #007bff';
              rect.style.backgroundColor = 'rgba(0, 123, 255, 0.1)';
              rect.style.pointerEvents = 'none'; 
              rect.style.zIndex = "10000"
              document.body.appendChild(rect);
          }

          
          document.addEventListener('mousedown', (e) => {
              console.log("mousedown: start", e); 
              startX = e.clientX;
              startY = e.clientY;
              createRect();
              rect.style.left = startX + 'px';
              rect.style.top = startY + 'px';
              rect.style.width = '0px';
              rect.style.height = '0px';

              
              body.style.pointerEvents = 'none';
              body.style.userSelect = 'none';
              console.log("Actions Blocked");
              console.log("mousedown: end", e); 
          });

          document.addEventListener('mousemove', (e) => {
              console.log("mousemove: start", e); 
              if (!startX || !startY || !rect) return;
              const width = Math.abs(e.clientX - startX);
              const height = Math.abs(e.clientY - startY);
              const left = Math.min(e.clientX, startX);
              const top = Math.min(e.clientY, startY);

              rect.style.left = left + 'px';
              rect.style.top = top + 'px';
              rect.style.width = width + 'px';
              rect.style.height = height + 'px';
              console.log("mousemove: end", e); 
          });

          document.addEventListener('mouseup', (e) => {
              console.log("mouseup: start", e);
              if (!startX || !startY || !rect) return;
              const width = Math.abs(e.clientX - startX);
              const height = Math.abs(e.clientY - startY);
              const left = Math.min(e.clientX, startX);
              const top = Math.min(e.clientY, startY);

             
              rect.remove();

              
              console.log("mouseup: Calling captureSelectedArea");
              captureSelectedArea(left, top, width, height);

              
              startX = null;
              startY = null;
              rect = null; 

              
              body.style.pointerEvents = 'auto'; 
              body.style.userSelect = 'auto'; 
              console.log("Actions Restored");
              console.log("mouseup: end", e); 
          });

          
          function captureSelectedArea(x, y, width, height) {
              console.log("captureSelectedArea called", x, y, width, height);
              console.log("captureSelectedArea is being called");
              console.log("captureSelectedArea: Sending 'captureVisibleTab' message to background");
              chrome.runtime.sendMessage({
                  action: 'captureVisibleTab',
                  x: x,
                  y: y,
                  width: width,
                  height: height
              });
          }
      }
  });
}


function sendScreenshotToServer(imageData, x, y, width, height) {
  console.log("sendScreenshotToServer called with data:", imageData, x, y, width, height);
  
  if (!imageData) {
      console.error("No image data to send.");
      return;
  }

  fetch('http://localhost:8080/upload', { 
    method: 'POST',
    headers: {
       'Content-Type': 'application/json' 
    },
    body: JSON.stringify({ image: imageData }) 
})
.then(response => {
    if (!response.ok) {
        console.error("HTTP error, status = " + response.status);
        return Promise.reject("HTTP error, status = " + response.status);
    }
    return response.json(); 
})
.then(data => {
    console.log("Success:", data); 
    
    if (data && data.message) {
        console.log("GPT Content:", data.message);
        
    } else {
        console.warn("No message received from server.");
    }
})
.catch(error => {
    console.error("Error sending screenshot:", error);
});
}

function startSelectionMode() {
  chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
      console.log("startSelectionMode: chrome.tabs.query result:", tabs); 
      if (!tabs || tabs.length === 0) {
          console.error("No active tab found in startSelectionMode.");
          return;
      }
      const tab = tabs[0];
      if (!tab) {
          console.error("No active tab found in startSelectionMode (after check).");
          return;
      }
      createSelectionRectangle(tab.id); 
  });
}

chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  console.log("onMessage: Message received", message);
  if (message.action === 'captureVisibleTab') {
      console.log("onMessage: captureVisibleTab action detected", message);
      chrome.tabs.captureVisibleTab(null, { format: 'jpeg', quality: 90 }, (screenshotUrl) => {
          if (chrome.runtime.lastError) {
              console.error("Screenshot capture error:", chrome.runtime.lastError);
              return;
          }

          console.log("Screenshot URL:", screenshotUrl);
          sendScreenshotToServer(screenshotUrl, message.x, message.y, message.width, message.height);
      });
  }
});

chrome.commands.onCommand.addListener((command) => {
  if (command === 'take-screenshot') {
      startSelectionMode();
  }
});