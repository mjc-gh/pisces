(function visibleText() {
	const isVisible = (element) => element.checkVisibility({
		opacityProperty: true,
		visibilityProperty: true,
		contentVisibilityAuto: true
	})

	// Create a NodeIterator to traverse all text nodes
	const nodeIterator = document.createNodeIterator(
		document.body, NodeFilter.SHOW_TEXT, {
			acceptNode: function(node) {
				// Check if the text node is visible
				if (!isVisible(node.parentElement)) {
					return NodeFilter.FILTER_REJECT;
				}

				// Check if it's a leaf text node (non-empty)
				const text = node.textContent.trim();
				if (text.length === 0) {
					return NodeFilter.FILTER_REJECT;
				}

				return NodeFilter.FILTER_ACCEPT;
			}
		}
	)

	// Iterate through all accepted text nodes
	const visibleTextParts = []

	let currentNode

	while (currentNode = nodeIterator.nextNode()) {
		visibleTextParts.push(currentNode.textContent.trim())
	}

	return visibleTextParts.join(" ")
}())
