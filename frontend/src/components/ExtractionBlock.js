import React, { useState } from 'react';

const API_URL = 'http://localhost:8080/api';

const ExtractionBlock = () => {
	const [selectedFile, setSelectedFile] = useState(undefined);
	const [fileUploaded, setFileUploaded] = useState(false);
	const [base64_image, setBase64_image] = useState("");

	const uploadImage = (e) => {
		e.preventDefault();

		console.log('Uploading image.');

		const formData = new FormData();
		formData.append('image', selectedFile);

		const options = {
			method: 'POST',
			headers: {
				'Content-Type': 'multipart/form-data',
				'Accept': 'application/json',
			},
			body: formData,
		}
		delete options.headers['Content-Type'];
		fetch(API_URL+'/upload', options)
		.then(response => {
			console.log(response);
			if (response.status === 200) {
				setFileUploaded(true);
			}
			return response.json();
		})
		.then(result => {
			console.log("Returned:", result);
			setBase64_image(result.base64_image);
		})
		.catch(error => {
			console.log('Upload failed.');
			console.error('Error:', error);
		});
	}

	return (
		<div className="extraction-block">
			<PaletteExtractForm 
				onSubmit={uploadImage} 
				selectedFile={selectedFile}
				setSelectedFile={setSelectedFile}
				fileUploaded={fileUploaded}
			/>
			{
				(base64_image === "")
					? <></>
					: <div
							className="uploaded-image"
						>
							<img
								src={base64_image}
								alt="Nothing to show"
								height="100%"
								width="100%"
							/>
						</div>
			}
		</div>
	);
}

const PaletteExtractForm = (props) => {
	return (
		<div>
			<form 
				onSubmit={(e) => props.onSubmit(e)}
				className="upload-form"
			>
				<div>Upload.</div>
				<input 
					type="file" 
					name="image"
					onChange={(e) => props.setSelectedFile(e.target.files[0])}
				/>
				<button>
					Upload image!
				</button>
			</form>
			<DisplayPalette 
				display={props.fileUploaded}
			/>
		</div>
	);
}

const DisplayPalette = ({ display }) => {
	const [numCols, setNumCols] = useState(3);
	const [numGo, setNumGo] = useState(4);
	const [tolerance, setTolerance] = useState(15);
	const [palette, setPalette] = useState([]);
	const [loading, setLoading] = useState(false);
	const [concurrent, setConcurrent] = useState(false);
	const [mode, setMode] = useState(0)

	if (!display) {
		return null;
	}

	const getPalette = () => {
		const urlToUse = 
		`${API_URL}/extract/?colors=${numCols}&concurrent=${concurrent}`
		+ `&mode=${mode}&goroutines=${numGo}&tolerance=${tolerance}`;

		const options = {
			method: 'GET',
		};
		console.log(`Sending this url: ${urlToUse}`);

		setLoading(true);
		fetch(urlToUse, options)
			.then(response => response.json())
			.then(result => {
				console.log('Returned:', result);
				setPalette(result);
				setLoading(false);
			})
			.catch(error => {
				console.log('Palette extraction failed.');
				console.error('Error:', error);
				setLoading(false);
			});
	}

	const options = [];
	for (let i = 3; i < 101; i++) {
		options.push(<option value={i} key={i}>{i}</option>);
	}

	const blocks = [];
	for (let i = 0; i < palette.length; i++) {
		blocks.push(
			<div
				className="color-block" 
			>
				<div 
					className="color-itself"
					style={{backgroundColor: palette[i].color}}
					key={i}
				>
				</div>
				<div
					className="color-info"
				>
					{palette[i].color}: {palette[i].frequency}
				</div>
			</div>
		);
	}

	return (
		<div className="find-palette">
			<div>
				<label htmlFor="extractNum">Colors to extract:</label>
				<select 
					id="extractNum"
					value={numCols}
					onChange={(e) => setNumCols(e.target.value)}
				>
					{options}
				</select>
				<input 
					type="radio" id="concurrent" name="concurrent" 
					onClick={() => setConcurrent(true)}
				/>
				<label for="concurrent">Concurrent</label>
				<input 
					type="radio" id="sequential" name="concurrent" 
					onClick={() => setConcurrent(false)}
					defaultChecked={!concurrent}
				/>
				<label for="sequential">Sequential</label>
				<button onClick={getPalette}>
					Find palette!
				</button>
			</div>
			<div>
				<label htmlFor="tolerance">Color tolerance:</label>
				<input
					id="tolerance"
					type="number"
					min="15"
					max="440"
					value={tolerance}
					onChange={(e) => setTolerance(e.target.value)}
				/>
			</div>
			{
				concurrent 
					? <div>
							<label htmlFor="goroutines">goroutines to use (4 - 100):</label>
							<input
								id="goroutines"
								type="number"
								min="4"
								max="100"
								value={numGo}
								onChange={(e) => setNumGo(e.target.value)}
							/>
						</div>
					: <></>
			}
			{loading 
				? <div>Loading...</div>
				: <div className="display-palette">
						{
							blocks
						}
					</div>
			}
		</div>
	);
}

export default ExtractionBlock;
